package resources

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
)

// These are obvious placeholders, not credentials. The tests search Terraform's
// human-readable plan, apply, state, refresh, import, and destroy output for
// them. They must be absent, and configuration must be shown as a sensitive
// value, while the same strings are still stored and sent to the API.
const (
	placeholderLDAP   = "PLACEHOLDER_CONFIGURATION_SECRET"
	placeholderSAML   = "PLACEHOLDER_SAML_SECRET"
	placeholderTwilio = "PLACEHOLDER_TWILIO_SECRET"
)

const (
	terraformVersion = "1.17.2"
	terraformZipSHA  = "e6f16b2f0b137ee5729caa8c4e79a3bae4ee261acd198a1fcfa8584a6296605b"
	terraformZipName = "terraform_1.17.2_linux_amd64.zip"
)

// TestAuthProviderConfigurationRedactedAcrossLifecycle drives the real
// ironwifi_authentication_provider resource through Terraform CLI against a
// local stand-in of the authentication-providers API. Create, apply, state
// show, update, refresh (Read), import (Read), and destroy plan must redact
// configuration. The API stand-in stores the configuration string and returns
// it unchanged, which is the value mapAuthProviderResponse writes into state.
func TestAuthProviderConfigurationRedactedAcrossLifecycle(t *testing.T) {
	terraformBin := terraformCLI(t)
	providerDir := buildProvider(t)
	api := newAuthProviderAPI(t)

	work := t.TempDir()
	writeFile(t, filepath.Join(work, "main.tf"), authProviderConfig(api.URL, placeholderLDAP, "ldap"))
	env := terraformEnv(t, providerDir)
	terraform(t, terraformBin, work, env, "init", "-no-color", "-input=false")

	planOut := terraform(t, terraformBin, work, env, "plan", "-no-color", "-input=false", "-lock=false")
	assertConfigurationRedacted(t, "create plan", planOut, placeholderLDAP)
	assertPlanMarksConfigurationSensitive(t, "create plan", terraform(t, terraformBin, work, env, "show", "-json", mustSavePlan(t, terraformBin, work, env)))

	applyOut := terraform(t, terraformBin, work, env, "apply", "-no-color", "-input=false", "-lock=false", "-auto-approve")
	assertConfigurationRedacted(t, "apply", applyOut, placeholderLDAP)
	if got := api.configurationOf(t, "ap-1"); !strings.Contains(got, placeholderLDAP) {
		t.Fatalf("create did not send configuration through to the authentication-providers API")
	}

	stateShow := terraform(t, terraformBin, work, env, "state", "show", "-no-color", "ironwifi_authentication_provider.ldap")
	assertConfigurationRedacted(t, "state show after apply", stateShow, placeholderLDAP)
	assertStateStoresAndMarksConfiguration(t, "state after apply", work, "ironwifi_authentication_provider.ldap", placeholderLDAP)

	writeFile(t, filepath.Join(work, "main.tf"), authProviderConfig(api.URL, placeholderSAML, "saml"))
	updatePlan := terraform(t, terraformBin, work, env, "plan", "-no-color", "-input=false", "-lock=false")
	assertConfigurationRedacted(t, "update plan", updatePlan, placeholderLDAP, placeholderSAML)
	updateApply := terraform(t, terraformBin, work, env, "apply", "-no-color", "-input=false", "-lock=false", "-auto-approve")
	assertConfigurationRedacted(t, "update apply", updateApply, placeholderLDAP, placeholderSAML)
	if got := api.configurationOf(t, "ap-1"); !strings.Contains(got, placeholderSAML) {
		t.Fatalf("update did not send configuration through to the authentication-providers API")
	}
	assertStateStoresAndMarksConfiguration(t, "state after update", work, "ironwifi_authentication_provider.ldap", placeholderSAML)

	api.setConfiguration(t, "ap-1", `{"twilio_token":"`+placeholderTwilio+`"}`)
	refreshOut := terraform(t, terraformBin, work, env, "apply", "-refresh-only", "-no-color", "-input=false", "-lock=false", "-auto-approve")
	assertConfigurationRedacted(t, "refresh apply", refreshOut, placeholderLDAP, placeholderSAML, placeholderTwilio)
	assertStateStoresAndMarksConfiguration(t, "state after refresh", work, "ironwifi_authentication_provider.ldap", placeholderTwilio)

	importDir := t.TempDir()
	writeFile(t, filepath.Join(importDir, "main.tf"), importedAuthProviderConfig(api.URL))
	terraform(t, terraformBin, importDir, env, "init", "-no-color", "-input=false")
	importOut := terraform(t, terraformBin, importDir, env, "import", "-no-color", "-input=false", "-lock=false", "ironwifi_authentication_provider.imported", "ap-1")
	assertPlaceholdersAbsent(t, "import", importOut, placeholderTwilio)
	importShow := terraform(t, terraformBin, importDir, env, "state", "show", "-no-color", "ironwifi_authentication_provider.imported")
	assertConfigurationRedacted(t, "state show after import", importShow, placeholderTwilio)
	assertStateStoresAndMarksConfiguration(t, "state after import", importDir, "ironwifi_authentication_provider.imported", placeholderTwilio)

	destroyPlan := terraform(t, terraformBin, work, env, "plan", "-destroy", "-no-color", "-input=false", "-lock=false")
	assertConfigurationRedacted(t, "destroy plan", destroyPlan, placeholderSAML, placeholderTwilio)
}

func authProviderConfig(endpoint, secret, providerType string) string {
	key := "dbpassword"
	if providerType == "saml" {
		key = "client_secret"
	}
	return fmt.Sprintf(`
terraform {
  required_providers {
    ironwifi = {
      source = "registry.terraform.io/ironwifi/ironwifi"
    }
  }
}

provider "ironwifi" {
  api_endpoint = %q
  api_token    = "placeholder-api-token"
  company_id   = "co"
}

resource "ironwifi_authentication_provider" "ldap" {
  name   = "Corporate LDAP"
  type   = %q
  status = "enabled"
  configuration = jsonencode({
    basedn = "dc=example,dc=com"
    bind   = "cn=admin,dc=example,dc=com"
    %s     = %q
  })
}
`, endpoint, providerType, key, secret)
}

func importedAuthProviderConfig(endpoint string) string {
	return fmt.Sprintf(`
terraform {
  required_providers {
    ironwifi = {
      source = "registry.terraform.io/ironwifi/ironwifi"
    }
  }
}

provider "ironwifi" {
  api_endpoint = %q
  api_token    = "placeholder-api-token"
  company_id   = "co"
}

resource "ironwifi_authentication_provider" "imported" {
  name   = "Corporate LDAP"
  type   = "ldap"
  status = "enabled"
}
`, endpoint)
}

func assertConfigurationRedacted(t *testing.T, label, output string, placeholders ...string) {
	t.Helper()
	assertPlaceholdersAbsent(t, label, output, placeholders...)
	if !regexp.MustCompile(`configuration\s*=\s*\(sensitive value\)`).MatchString(output) {
		t.Errorf("%s did not render configuration as a sensitive value", label)
	}
}

func assertPlaceholdersAbsent(t *testing.T, label, output string, placeholders ...string) {
	t.Helper()
	for _, placeholder := range placeholders {
		if strings.Contains(output, placeholder) {
			t.Errorf("%s printed the configuration placeholder in Terraform output", label)
		}
	}
}

func assertPlanMarksConfigurationSensitive(t *testing.T, label, showJSON string) {
	t.Helper()
	var plan struct {
		ResourceChanges []struct {
			Address string `json:"address"`
			Change  struct {
				AfterSensitive json.RawMessage `json:"after_sensitive"`
			} `json:"change"`
		} `json:"resource_changes"`
	}
	if err := json.Unmarshal([]byte(showJSON), &plan); err != nil {
		t.Fatalf("%s: parsing plan JSON: %v", label, err)
	}
	for _, change := range plan.ResourceChanges {
		if change.Address != "ironwifi_authentication_provider.ldap" {
			continue
		}
		var sensitive map[string]any
		if err := json.Unmarshal(change.Change.AfterSensitive, &sensitive); err != nil {
			t.Fatalf("%s: parsing after_sensitive: %v", label, err)
		}
		marked, _ := sensitive["configuration"].(bool)
		if !marked {
			t.Errorf("%s JSON plan does not mark configuration sensitive", label)
		}
		return
	}
	t.Errorf("%s JSON plan has no authentication provider change", label)
}

func assertStateStoresAndMarksConfiguration(t *testing.T, label, workdir, address, placeholder string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(workdir, "terraform.tfstate"))
	if err != nil {
		t.Fatalf("%s: reading state: %v", label, err)
	}
	var state struct {
		Resources []struct {
			Type      string `json:"type"`
			Name      string `json:"name"`
			Instances []struct {
				AttributesRaw           json.RawMessage `json:"attributes"`
				AttributeSensitivePaths json.RawMessage `json:"sensitive_attributes"`
			} `json:"instances"`
		} `json:"resources"`
	}
	if err := json.Unmarshal(raw, &state); err != nil {
		t.Fatalf("%s: parsing state: %v", label, err)
	}
	wantType, wantName, ok := strings.Cut(strings.TrimPrefix(address, "ironwifi_"), ".")
	if !ok {
		t.Fatalf("bad address %s", address)
	}
	wantType = "ironwifi_" + wantType
	for _, resource := range state.Resources {
		if resource.Type != wantType || resource.Name != wantName || len(resource.Instances) == 0 {
			continue
		}
		inst := resource.Instances[0]
		var attrs map[string]any
		if err := json.Unmarshal(inst.AttributesRaw, &attrs); err != nil {
			t.Fatalf("%s: parsing attributes: %v", label, err)
		}
		configuration, _ := attrs["configuration"].(string)
		if !strings.Contains(configuration, placeholder) {
			t.Errorf("%s does not store the configuration returned by the API", label)
		}
		if !bytes.Contains(inst.AttributeSensitivePaths, []byte("configuration")) {
			t.Errorf("%s does not record configuration in sensitive_attributes", label)
		}
		return
	}
	t.Errorf("%s is missing %s", label, address)
}

func mustSavePlan(t *testing.T, bin, dir string, env []string) string {
	t.Helper()
	planPath := filepath.Join(dir, "create.tfplan")
	terraform(t, bin, dir, env, "plan", "-no-color", "-input=false", "-lock=false", "-out", planPath)
	return planPath
}

func terraform(t *testing.T, bin, dir string, env []string, args ...string) string {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("terraform %s failed: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func terraformEnv(t *testing.T, providerDir string) []string {
	t.Helper()
	cli := filepath.Join(t.TempDir(), "terraform.rc")
	writeFile(t, cli, fmt.Sprintf(`
provider_installation {
  dev_overrides {
    "registry.terraform.io/ironwifi/ironwifi" = %q
  }
  direct {}
}
`, providerDir))
	env := make([]string, 0, len(os.Environ())+3)
	for _, entry := range os.Environ() {
		if strings.HasPrefix(entry, "TF_LOG") || strings.HasPrefix(entry, "TF_CLI_CONFIG_FILE=") {
			continue
		}
		env = append(env, entry)
	}
	env = append(env,
		"TF_CLI_CONFIG_FILE="+cli,
		"TF_IN_AUTOMATION=1",
		"CHECKPOINT_DISABLE=1",
	)
	return env
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

var (
	providerBuildOnce sync.Once
	providerBuildDir  string
	providerBuildErr  error
	terraformOnce     sync.Once
	terraformPath     string
	terraformErr      error
)

func buildProvider(t *testing.T) string {
	t.Helper()
	providerBuildOnce.Do(func() {
		dir, err := os.MkdirTemp("", "ironwifi-provider-")
		if err != nil {
			providerBuildErr = err
			return
		}
		bin := filepath.Join(dir, "terraform-provider-ironwifi")
		cmd := exec.Command("go", "build", "-o", bin, ".")
		cmd.Dir = moduleRoot(t)
		out, err := cmd.CombinedOutput()
		if err != nil {
			providerBuildErr = fmt.Errorf("go build: %w\n%s", err, out)
			return
		}
		providerBuildDir = dir
	})
	if providerBuildErr != nil {
		t.Fatalf("building provider: %v", providerBuildErr)
	}
	return providerBuildDir
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func terraformCLI(t *testing.T) string {
	t.Helper()
	terraformOnce.Do(func() {
		terraformPath, terraformErr = ensureTerraform()
	})
	if terraformErr != nil {
		t.Fatalf("terraform CLI: %v", terraformErr)
	}
	return terraformPath
}

func ensureTerraform() (string, error) {
	cache := filepath.Join(os.TempDir(), "terraform-"+terraformVersion+"-linux-amd64")
	bin := filepath.Join(cache, "terraform")
	if info, err := os.Stat(bin); err == nil && !info.IsDir() {
		return bin, nil
	}
	if err := os.MkdirAll(cache, 0o755); err != nil {
		return "", err
	}
	url := "https://releases.hashicorp.com/terraform/" + terraformVersion + "/" + terraformZipName
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("downloading terraform: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(body)
	if hex.EncodeToString(sum[:]) != terraformZipSHA {
		return "", fmt.Errorf("terraform zip checksum mismatch")
	}
	reader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return "", err
	}
	for _, file := range reader.File {
		if file.Name != "terraform" {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()
		out, err := os.OpenFile(bin, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			return "", err
		}
		if err := out.Close(); err != nil {
			return "", err
		}
		return bin, nil
	}
	return "", fmt.Errorf("terraform binary missing from zip")
}

type authProviderAPI struct {
	URL    string
	server *httptest.Server
	mu     sync.Mutex
	items  map[string]map[string]any
	nextID int
}

func newAuthProviderAPI(t *testing.T) *authProviderAPI {
	t.Helper()
	api := &authProviderAPI{items: map[string]map[string]any{}, nextID: 1}
	api.server = httptest.NewServer(http.HandlerFunc(api.serve))
	api.URL = api.server.URL
	t.Cleanup(api.server.Close)
	return api
}

func (a *authProviderAPI) serve(w http.ResponseWriter, r *http.Request) {
	const prefix = "/api/co/authentication-providers"
	if r.URL.Path != prefix && !strings.HasPrefix(r.URL.Path, prefix+"/") {
		http.NotFound(w, r)
		return
	}
	id := strings.TrimPrefix(strings.TrimPrefix(r.URL.Path, prefix), "/")
	switch {
	case r.Method == http.MethodPost && id == "":
		body := a.decode(w, r)
		if body == nil {
			return
		}
		a.mu.Lock()
		itemID := fmt.Sprintf("ap-%d", a.nextID)
		a.nextID++
		body["id"] = itemID
		a.items[itemID] = body
		a.mu.Unlock()
		writeJSON(w, body)
	case r.Method == http.MethodGet && id != "":
		a.mu.Lock()
		item, ok := a.items[id]
		a.mu.Unlock()
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, item)
	case r.Method == http.MethodPatch && id != "":
		body := a.decode(w, r)
		if body == nil {
			return
		}
		a.mu.Lock()
		item, ok := a.items[id]
		if ok {
			for k, v := range body {
				item[k] = v
			}
			item["id"] = id
		}
		a.mu.Unlock()
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, item)
	case r.Method == http.MethodDelete && id != "":
		a.mu.Lock()
		delete(a.items, id)
		a.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "unsupported", http.StatusMethodNotAllowed)
	}
}

func (a *authProviderAPI) decode(w http.ResponseWriter, r *http.Request) map[string]any {
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil
	}
	return body
}

func writeJSON(w http.ResponseWriter, body map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}

func (a *authProviderAPI) configurationOf(t *testing.T, id string) string {
	t.Helper()
	a.mu.Lock()
	defer a.mu.Unlock()
	item, ok := a.items[id]
	if !ok {
		t.Fatalf("api has no authentication provider %s", id)
	}
	configuration, _ := item["configuration"].(string)
	return configuration
}

func (a *authProviderAPI) setConfiguration(t *testing.T, id, configuration string) {
	t.Helper()
	a.mu.Lock()
	defer a.mu.Unlock()
	item, ok := a.items[id]
	if !ok {
		t.Fatalf("api has no authentication provider %s", id)
	}
	item["configuration"] = configuration
}
