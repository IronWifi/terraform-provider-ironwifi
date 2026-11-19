package resources

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ironwifi/terraform-provider-ironwifi/internal/client"
)

// Obvious placeholders standing in for the two password values an operator
// would put in configuration. They are not credentials.
const (
	userPasswordOriginal = "placeholder-original"
	userPasswordRotated  = "placeholder-rotated"
)

// TestUserPasswordUpdateRecordsOnlyAPasswordTheAPIWasSent is the spec for
// ironwifi_user password changes. Create sends the password. Update must do
// the same whenever the planned password is set, and the password written to
// state must be that same value. A successful update that stores a planned
// password the users request never carried leaves the previous password in
// place while the next plan is empty.
func TestUserPasswordUpdateRecordsOnlyAPasswordTheAPIWasSent(t *testing.T) {
	t.Run("password change", func(t *testing.T) {
		original := testUserModel("", "ada@example.com", userPasswordOriginal, "Ada")
		rotated := original
		rotated.Password = types.StringValue(userPasswordRotated)
		rotated.Firstname = types.StringValue("Grace")
		assertUserPasswordUpdateSendsPlannedPassword(t, original, rotated)
	})

	t.Run("password set while another attribute changes", func(t *testing.T) {
		original := testUserModel("", "ada@example.com", userPasswordOriginal, "Ada")
		updated := original
		updated.Firstname = types.StringValue("Grace")
		assertUserPasswordUpdateSendsPlannedPassword(t, original, updated)
	})
}

func assertUserPasswordUpdateSendsPlannedPassword(t *testing.T, createModel, updatePlan UserResourceModel) {
	t.Helper()
	ctx := context.Background()

	api := newUsersAPIStub()
	srv := httptest.NewServer(api)
	t.Cleanup(srv.Close)

	c, err := client.New(&client.Config{
		APIEndpoint: srv.URL,
		APIToken:    "placeholder-api-token",
		CompanyID:   "company-1",
	})
	if err != nil {
		t.Fatalf("client: %v", err)
	}

	r := &UserResource{}
	configureResp := &resource.ConfigureResponse{}
	r.Configure(ctx, resource.ConfigureRequest{ProviderData: c}, configureResp)
	if configureResp.Diagnostics.HasError() {
		t.Fatalf("configure: %s", diagnosticsSummary(configureResp.Diagnostics))
	}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	userSchema := schemaResp.Schema

	createPlan := userSchemaPlan(t, userSchema, createModel)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Schema: userSchema},
	}
	r.Create(ctx, resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("create: %s", diagnosticsSummary(createResp.Diagnostics))
	}

	var created UserResourceModel
	createResp.State.Get(ctx, &created)
	if created.ID.ValueString() == "" {
		t.Fatal("create did not store a user id")
	}
	if created.Password.ValueString() != createModel.Password.ValueString() {
		t.Fatal("create state password differs from the planned password")
	}

	createCall := api.callsWith("POST")
	if len(createCall) != 1 {
		t.Fatalf("create issued %d POST requests", len(createCall))
	}
	if _, ok := createCall[0].body["password"]; !ok {
		t.Fatal("create request omitted password")
	}
	if createCall[0].body["password"] != createModel.Password.ValueString() {
		t.Fatal("create request password differs from the planned password")
	}

	updatePlan.ID = created.ID
	updatePlan.CreationDate = created.CreationDate
	prior := created
	plan := userSchemaPlan(t, userSchema, updatePlan)
	state := userSchemaState(t, userSchema, prior)

	updateResp := &resource.UpdateResponse{State: state}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: state}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("update: %s", diagnosticsSummary(updateResp.Diagnostics))
	}

	var updated UserResourceModel
	updateResp.State.Get(ctx, &updated)
	if updated.Password.ValueString() != updatePlan.Password.ValueString() {
		t.Fatal("update state password differs from the planned password")
	}

	patches := api.callsWith("PATCH")
	if len(patches) != 1 {
		t.Fatalf("update issued %d PATCH requests", len(patches))
	}
	patch := patches[0]
	if patch.path != "/api/company-1/users/"+created.ID.ValueString() {
		t.Fatalf("update path = %s", patch.path)
	}
	if patch.body["username"] != updatePlan.Username.ValueString() {
		t.Fatal("update request did not send the planned username")
	}
	if patch.body["firstname"] != updatePlan.Firstname.ValueString() {
		t.Fatal("update request did not send the planned firstname")
	}
	sent, ok := patch.body["password"]
	if !ok {
		t.Fatal("update request omitted password while the planned password is set and the apply stored it in state")
	}
	if sent != updatePlan.Password.ValueString() {
		t.Fatal("update request password differs from the planned password stored in state")
	}

	// Read must not hide a password the update request never carried. The
	// users API response has no password field; Read keeps whatever Update
	// wrote, which is what the next plan compares against configuration.
	readState := userSchemaState(t, userSchema, updated)
	readResp := &resource.ReadResponse{State: readState}
	r.Read(ctx, resource.ReadRequest{State: readState}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("read: %s", diagnosticsSummary(readResp.Diagnostics))
	}
	var refreshed UserResourceModel
	readResp.State.Get(ctx, &refreshed)
	if refreshed.Password.ValueString() != updatePlan.Password.ValueString() {
		t.Fatal("read changed the password stored in state")
	}
	if refreshed.Password.ValueString() != api.passwordSentFor(created.ID.ValueString()) {
		t.Fatal("state password differs from the password last sent to the users API, so the next plan has nothing to correct")
	}
}

func testUserModel(id, username, password, firstname string) UserResourceModel {
	idValue := types.StringNull()
	if id != "" {
		idValue = types.StringValue(id)
	}
	return UserResourceModel{
		ID:           idValue,
		Username:     types.StringValue(username),
		Password:     types.StringValue(password),
		Email:        types.StringValue(""),
		Firstname:    types.StringValue(firstname),
		Lastname:     types.StringValue(""),
		Notes:        types.StringValue(""),
		UserType:     types.StringValue("e"),
		MobilePhone:  types.StringValue(""),
		AuthSource:   types.StringValue("local"),
		OrgUnit:      types.StringValue(""),
		Status:       types.StringValue(""),
		DeletionDate: types.StringValue(""),
		CreationDate: types.StringNull(),
	}
}

func userSchemaPlan(t *testing.T, userSchema schema.Schema, model UserResourceModel) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: userSchema}
	diags := plan.Set(context.Background(), &model)
	if diags.HasError() {
		t.Fatalf("plan: %s", diagnosticsSummary(diags))
	}
	return plan
}

func userSchemaState(t *testing.T, userSchema schema.Schema, model UserResourceModel) tfsdk.State {
	t.Helper()
	state := tfsdk.State{Schema: userSchema}
	diags := state.Set(context.Background(), &model)
	if diags.HasError() {
		t.Fatalf("state: %s", diagnosticsSummary(diags))
	}
	return state
}

func diagnosticsSummary(diags diag.Diagnostics) string {
	parts := make([]string, 0, len(diags))
	for _, d := range diags.Errors() {
		parts = append(parts, d.Summary())
	}
	return strings.Join(parts, "; ")
}

// usersAPIStub is the users collection the provider client talks to.
// It records each JSON body and remembers a password only when that body
// contains one. Responses omit password, matching the resource's Read path,
// which keeps the state value because the response has no password to copy.
type usersAPIStub struct {
	mu    sync.Mutex
	users map[string]storedUser
	calls []recordedCall
}

type storedUser struct {
	public   map[string]interface{}
	password string
	has      bool
}

type recordedCall struct {
	method string
	path   string
	body   map[string]interface{}
}

func newUsersAPIStub() *usersAPIStub {
	return &usersAPIStub{users: map[string]storedUser{}}
}

func (s *usersAPIStub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read", http.StatusBadRequest)
		return
	}

	var body map[string]interface{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &body); err != nil {
			http.Error(w, "json", http.StatusBadRequest)
			return
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, recordedCall{method: r.Method, path: r.URL.Path, body: body})

	w.Header().Set("Content-Type", "application/hal+json")

	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/api/company-1/users":
		id := "user-1"
		s.users[id] = userFromBody(id, body, storedUser{})
		writeUser(w, s.users[id])
	case r.Method == http.MethodPatch && strings.HasPrefix(r.URL.Path, "/api/company-1/users/"):
		id := strings.TrimPrefix(r.URL.Path, "/api/company-1/users/")
		current, ok := s.users[id]
		if !ok {
			http.NotFound(w, r)
			return
		}
		s.users[id] = userFromBody(id, body, current)
		writeUser(w, s.users[id])
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/company-1/users/"):
		id := strings.TrimPrefix(r.URL.Path, "/api/company-1/users/")
		current, ok := s.users[id]
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeUser(w, current)
	default:
		http.NotFound(w, r)
	}
}

func userFromBody(id string, body map[string]interface{}, current storedUser) storedUser {
	if current.public == nil {
		current.public = map[string]interface{}{
			"id":           id,
			"creationdate": "2024-01-01 00:00:00",
		}
	}
	for key, value := range body {
		if key == "password" {
			current.password = value.(string)
			current.has = true
			continue
		}
		current.public[key] = value
	}
	current.public["id"] = id
	return current
}

func writeUser(w http.ResponseWriter, user storedUser) {
	payload := map[string]interface{}{}
	for key, value := range user.public {
		payload[key] = value
	}
	delete(payload, "password")
	_ = json.NewEncoder(w).Encode(payload)
}

func (s *usersAPIStub) callsWith(method string) []recordedCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	var matched []recordedCall
	for _, call := range s.calls {
		if call.method == method {
			matched = append(matched, call)
		}
	}
	return matched
}

func (s *usersAPIStub) passwordSentFor(id string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[id]
	if !ok || !user.has {
		return ""
	}
	return user.password
}
