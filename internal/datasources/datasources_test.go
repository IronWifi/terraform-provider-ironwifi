package datasources

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ironwifi/terraform-provider-ironwifi/internal/client"
)

func testClient(t *testing.T, handler http.HandlerFunc) *client.Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	c, err := client.New(&client.Config{
		APIEndpoint: server.URL,
		CompanyID:   "test-company",
		APIToken:    "test-token",
	})
	if err != nil {
		t.Fatalf("creating test client: %v", err)
	}
	return c
}

func TestNetworksDataSource_List(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"_embedded": map[string]interface{}{
				"networks": []map[string]interface{}{
					{"id": "n1", "nasname": "OfficeWiFi", "region": "us-east1", "auth_port": float64(1812), "acct_port": float64(1813), "primary_ip": "10.0.0.1", "secret": "s3cret"},
					{"id": "n2", "nasname": "GuestWiFi", "region": "eu-west1", "auth_port": float64(1812), "acct_port": float64(1813), "primary_ip": "10.0.0.2", "secret": "guest"},
				},
			},
			"total_items": 2,
		})
	})

	items, err := c.List("networks", "networks")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0]["nasname"] != "OfficeWiFi" {
		t.Errorf("expected OfficeWiFi, got %v", items[0]["nasname"])
	}
}

func TestUsersDataSource_List(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"_embedded": map[string]interface{}{
				"users": []map[string]interface{}{
					{"id": "u1", "username": "admin@example.com", "email": "admin@example.com", "firstname": "Admin", "lastname": "User"},
				},
			},
			"total_items": 1,
		})
	})

	items, err := c.List("users", "users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0]["username"] != "admin@example.com" {
		t.Errorf("expected admin@example.com, got %v", items[0]["username"])
	}
}

func TestGroupsDataSource_List(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"_embedded": map[string]interface{}{
				"groups": []map[string]interface{}{
					{"id": "g1", "groupname": "Engineering", "description": "Eng team", "priority": float64(10)},
					{"id": "g2", "groupname": "Marketing", "description": "", "priority": float64(20)},
				},
			},
			"total_items": 2,
		})
	})

	items, err := c.List("groups", "groups")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestPoliciesDataSource_List(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"_embedded": map[string]interface{}{
				"conditional_access_policies": []map[string]interface{}{
					{"id": "p1", "name": "Bandwidth Limit", "enabled": float64(1), "priority": float64(50)},
				},
			},
			"total_items": 1,
		})
	})

	items, err := c.List("policies", "conditional_access_policies")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

func TestAuthProvidersDataSource_List(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"_embedded": map[string]interface{}{
				"authentication_providers": []map[string]interface{}{
					{"id": "ap1", "name": "LDAP", "type": "ldap", "status": "enabled"},
				},
			},
			"total_items": 1,
		})
	})

	items, err := c.List("authentication-providers", "authentication_providers")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

func TestHelpers_StringVal(t *testing.T) {
	data := map[string]interface{}{"name": "test", "nil_val": nil}

	v := stringVal(data, "name")
	if v.ValueString() != "test" {
		t.Errorf("expected 'test', got %q", v.ValueString())
	}

	v2 := stringVal(data, "missing")
	if v2.ValueString() != "" {
		t.Errorf("expected empty string, got %q", v2.ValueString())
	}

	v3 := stringVal(data, "nil_val")
	if v3.ValueString() != "" {
		t.Errorf("expected empty string for nil, got %q", v3.ValueString())
	}
}

func TestHelpers_IntVal(t *testing.T) {
	data := map[string]interface{}{"port": float64(1812)}
	v := intVal(data, "port")
	if v.ValueInt64() != 1812 {
		t.Errorf("expected 1812, got %d", v.ValueInt64())
	}

	v2 := intVal(data, "missing")
	if v2.ValueInt64() != 0 {
		t.Errorf("expected 0, got %d", v2.ValueInt64())
	}
}

func TestHelpers_BoolVal(t *testing.T) {
	data := map[string]interface{}{"enabled": true, "disabled": false, "int_true": float64(1), "int_false": float64(0)}

	if !boolVal(data, "enabled").ValueBool() {
		t.Error("expected true")
	}
	if boolVal(data, "disabled").ValueBool() {
		t.Error("expected false")
	}
	if !boolVal(data, "int_true").ValueBool() {
		t.Error("expected true for float64(1)")
	}
	if boolVal(data, "int_false").ValueBool() {
		t.Error("expected false for float64(0)")
	}
	if boolVal(data, "missing").ValueBool() {
		t.Error("expected false for missing")
	}
}

func TestEmptyEmbedded(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total_items": 0,
		})
	})

	items, err := c.List("networks", "networks")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}
