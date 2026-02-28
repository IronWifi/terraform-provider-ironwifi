package resources

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMapNetworkResponse(t *testing.T) {
	data := map[string]interface{}{
		"id":            "net-123",
		"nasname":       "TestNetwork",
		"region":        "us-east1",
		"auth_port":     float64(1812),
		"acct_port":     float64(1813),
		"secret":        "shared-secret",
		"primary_ip":    "10.0.0.1",
		"backup_ip":     "10.0.0.2",
		"ipv6":          float64(1),
		"unknown_users": "reject",
		"open_roaming":  float64(0),
		"eduroam":       float64(0),
		"coa":           float64(1),
		"radsec":        float64(0),
	}

	var model NetworkResourceModel
	mapNetworkResponse(data, &model)

	assertString(t, "ID", model.ID, "net-123")
	assertString(t, "Name", model.Name, "TestNetwork")
	assertString(t, "Region", model.Region, "us-east1")
	assertInt64(t, "AuthPort", model.AuthPort, 1812)
	assertInt64(t, "AcctPort", model.AcctPort, 1813)
	assertString(t, "Secret", model.Secret, "shared-secret")
	assertString(t, "PrimaryIP", model.PrimaryIP, "10.0.0.1")
	assertString(t, "BackupIP", model.BackupIP, "10.0.0.2")
	assertBool(t, "IPv6", model.IPv6, true)
	assertString(t, "UnknownUsers", model.UnknownUsers, "reject")
	assertBool(t, "OpenRoaming", model.OpenRoaming, false)
	assertBool(t, "Eduroam", model.Eduroam, false)
	assertBool(t, "COA", model.COA, true)
	assertBool(t, "RadSec", model.RadSec, false)
}

func TestSetBoolAsInt(t *testing.T) {
	body := make(map[string]interface{})

	setBoolAsInt(body, "enabled", types.BoolValue(true))
	if body["enabled"] != 1 {
		t.Errorf("expected 1 for true, got %v", body["enabled"])
	}

	setBoolAsInt(body, "disabled", types.BoolValue(false))
	if body["disabled"] != 0 {
		t.Errorf("expected 0 for false, got %v", body["disabled"])
	}

	setBoolAsInt(body, "null", types.BoolNull())
	if _, ok := body["null"]; ok {
		t.Error("null value should not be set")
	}
}

func TestBoolFromIntAPI(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		expected bool
	}{
		{"float64 1", map[string]interface{}{"k": float64(1)}, "k", true},
		{"float64 0", map[string]interface{}{"k": float64(0)}, "k", false},
		{"bool true", map[string]interface{}{"k": true}, "k", true},
		{"bool false", map[string]interface{}{"k": false}, "k", false},
		{"string 1", map[string]interface{}{"k": "1"}, "k", true},
		{"string 0", map[string]interface{}{"k": "0"}, "k", false},
		{"missing", map[string]interface{}{}, "k", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := boolFromIntAPI(tt.data, tt.key)
			if result.ValueBool() != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result.ValueBool())
			}
		})
	}
}

func TestStringFromAPI(t *testing.T) {
	data := map[string]interface{}{"name": "test", "empty": ""}
	if v := stringFromAPI(data, "name"); v.ValueString() != "test" {
		t.Errorf("expected 'test', got '%s'", v.ValueString())
	}
	if v := stringFromAPI(data, "missing"); v.ValueString() != "" {
		t.Errorf("expected empty string, got '%s'", v.ValueString())
	}
}

func TestIntFromAPI(t *testing.T) {
	data := map[string]interface{}{"port": float64(1812), "zero": float64(0)}
	if v := intFromAPI(data, "port"); v.ValueInt64() != 1812 {
		t.Errorf("expected 1812, got %d", v.ValueInt64())
	}
	if v := intFromAPI(data, "missing"); v.ValueInt64() != 0 {
		t.Errorf("expected 0, got %d", v.ValueInt64())
	}
}

// Test helpers
func assertString(t *testing.T, field string, got types.String, expected string) {
	t.Helper()
	if got.ValueString() != expected {
		t.Errorf("%s: expected %q, got %q", field, expected, got.ValueString())
	}
}

func assertInt64(t *testing.T, field string, got types.Int64, expected int64) {
	t.Helper()
	if got.ValueInt64() != expected {
		t.Errorf("%s: expected %d, got %d", field, expected, got.ValueInt64())
	}
}

func assertBool(t *testing.T, field string, got types.Bool, expected bool) {
	t.Helper()
	if got.ValueBool() != expected {
		t.Errorf("%s: expected %v, got %v", field, expected, got.ValueBool())
	}
}
