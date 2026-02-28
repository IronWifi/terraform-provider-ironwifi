package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ironwifi/terraform-provider-ironwifi/internal/client"
)

var (
	_ resource.Resource                = &UserResource{}
	_ resource.ResourceWithImportState = &UserResource{}
)

type UserResource struct {
	client *client.Client
}

type UserResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
	Email        types.String `tfsdk:"email"`
	Firstname    types.String `tfsdk:"firstname"`
	Lastname     types.String `tfsdk:"lastname"`
	Notes        types.String `tfsdk:"notes"`
	UserType     types.String `tfsdk:"user_type"`
	MobilePhone  types.String `tfsdk:"mobilephone"`
	AuthSource   types.String `tfsdk:"authsource"`
	OrgUnit      types.String `tfsdk:"orgunit"`
	Status       types.String `tfsdk:"status"`
	DeletionDate types.String `tfsdk:"deletiondate"`
	CreationDate types.String `tfsdk:"creationdate"`
}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

func (r *UserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an IronWiFi user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "User ID (UUID).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Description: "Username for authentication.",
				Required:    true,
			},
			"password": schema.StringAttribute{
				Description: "User password. Only sent on create.",
				Optional:    true,
				Sensitive:   true,
			},
			"email": schema.StringAttribute{
				Description: "User email address.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"firstname": schema.StringAttribute{
				Description: "User first name.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"lastname": schema.StringAttribute{
				Description: "User last name.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"notes": schema.StringAttribute{
				Description: "Notes about the user.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"user_type": schema.StringAttribute{
				Description: "User type: 'e' for employee, 'u' for user.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("e"),
			},
			"mobilephone": schema.StringAttribute{
				Description: "User mobile phone number.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"authsource": schema.StringAttribute{
				Description: "Authentication source (e.g. 'local', 'ldap').",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("local"),
			},
			"orgunit": schema.StringAttribute{
				Description: "Organizational unit.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"status": schema.StringAttribute{
				Description: "User status.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"deletiondate": schema.StringAttribute{
				Description: "Expiration/deletion date for the user.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"creationdate": schema.StringAttribute{
				Description: "Date the user was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *UserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"username": plan.Username.ValueString(),
	}
	setIfNotNull(body, "password", plan.Password)
	setIfNotNull(body, "email", plan.Email)
	setIfNotNull(body, "firstname", plan.Firstname)
	setIfNotNull(body, "lastname", plan.Lastname)
	setIfNotNull(body, "notes", plan.Notes)
	setIfNotNull(body, "user_type", plan.UserType)
	setIfNotNull(body, "mobilephone", plan.MobilePhone)
	setIfNotNull(body, "authsource", plan.AuthSource)
	setIfNotNull(body, "orgunit", plan.OrgUnit)
	setIfNotNull(body, "status", plan.Status)
	setIfNotNull(body, "deletiondate", plan.DeletionDate)

	result, err := r.client.Create("users", body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating user", err.Error())
		return
	}

	mapUserResponse(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Read("users", state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading user", err.Error())
		return
	}

	// Preserve password since the API does not return it.
	pw := state.Password
	mapUserResponse(result, &state)
	state.Password = pw
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan UserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"username": plan.Username.ValueString(),
	}
	setIfNotNull(body, "email", plan.Email)
	setIfNotNull(body, "firstname", plan.Firstname)
	setIfNotNull(body, "lastname", plan.Lastname)
	setIfNotNull(body, "notes", plan.Notes)
	setIfNotNull(body, "user_type", plan.UserType)
	setIfNotNull(body, "mobilephone", plan.MobilePhone)
	setIfNotNull(body, "authsource", plan.AuthSource)
	setIfNotNull(body, "orgunit", plan.OrgUnit)
	setIfNotNull(body, "status", plan.Status)
	setIfNotNull(body, "deletiondate", plan.DeletionDate)

	result, err := r.client.Update("users", state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating user", err.Error())
		return
	}

	// Preserve password since the API does not return it.
	pw := plan.Password
	mapUserResponse(result, &plan)
	plan.Password = pw
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete("users", state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting user", err.Error())
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapUserResponse(data map[string]interface{}, model *UserResourceModel) {
	model.ID = stringFromAPI(data, "id")
	model.Username = stringFromAPI(data, "username")
	model.Email = stringFromAPI(data, "email")
	model.Firstname = stringFromAPI(data, "firstname")
	model.Lastname = stringFromAPI(data, "lastname")
	model.Notes = stringFromAPI(data, "notes")
	model.UserType = stringFromAPI(data, "user_type")
	model.MobilePhone = stringFromAPI(data, "mobilephone")
	model.AuthSource = stringFromAPI(data, "authsource")
	model.OrgUnit = stringFromAPI(data, "orgunit")
	model.Status = stringFromAPI(data, "status")
	model.DeletionDate = stringFromAPI(data, "deletiondate")
	model.CreationDate = stringFromAPI(data, "creationdate")
}
