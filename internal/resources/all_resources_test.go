package resources

import (
	"testing"
)

// --- User resource mapping tests ---

func TestMapUserResponse(t *testing.T) {
	data := map[string]interface{}{
		"id":           "user-123",
		"username":     "john@example.com",
		"email":        "john@example.com",
		"firstname":    "John",
		"lastname":     "Doe",
		"notes":        "test user",
		"user_type":    "e",
		"mobilephone":  "+1234567890",
		"authsource":   "local",
		"orgunit":      "Engineering",
		"status":       "active",
		"deletiondate": "",
		"creationdate": "2024-01-01 00:00:00",
	}

	var model UserResourceModel
	mapUserResponse(data, &model)

	assertString(t, "ID", model.ID, "user-123")
	assertString(t, "Username", model.Username, "john@example.com")
	assertString(t, "Email", model.Email, "john@example.com")
	assertString(t, "Firstname", model.Firstname, "John")
	assertString(t, "Lastname", model.Lastname, "Doe")
	assertString(t, "Notes", model.Notes, "test user")
	assertString(t, "UserType", model.UserType, "e")
	assertString(t, "MobilePhone", model.MobilePhone, "+1234567890")
	assertString(t, "AuthSource", model.AuthSource, "local")
	assertString(t, "OrgUnit", model.OrgUnit, "Engineering")
	assertString(t, "Status", model.Status, "active")
	assertString(t, "CreationDate", model.CreationDate, "2024-01-01 00:00:00")
}

func TestMapUserResponse_PasswordNotOverwritten(t *testing.T) {
	// mapUserResponse does NOT set Password — it's write-only.
	// The Read handler must preserve it from state.
	data := map[string]interface{}{"id": "u1", "username": "u"}
	var model UserResourceModel
	mapUserResponse(data, &model)
	if !model.Password.IsNull() {
		t.Error("Password should remain null after mapping — it's write-only")
	}
}

// --- Group resource mapping tests ---

func TestMapGroupResponse(t *testing.T) {
	data := map[string]interface{}{
		"id":          "grp-1",
		"groupname":   "Engineering",
		"description": "Engineering team",
		"priority":    float64(10),
	}

	var model GroupResourceModel
	mapGroupResponse(data, &model)

	assertString(t, "ID", model.ID, "grp-1")
	assertString(t, "Name", model.Name, "Engineering")
	assertString(t, "Description", model.Description, "Engineering team")
	assertInt64(t, "Priority", model.Priority, 10)
}

func TestMapGroupResponse_Defaults(t *testing.T) {
	data := map[string]interface{}{"id": "g1", "groupname": "test"}
	var model GroupResourceModel
	mapGroupResponse(data, &model)
	assertInt64(t, "Priority", model.Priority, 0) // missing = 0
}

// --- Policy resource mapping tests ---

func TestMapPolicyResponse(t *testing.T) {
	data := map[string]interface{}{
		"id":          "pol-1",
		"name":        "BandwidthLimit",
		"description": "Limits bandwidth",
		"priority":    float64(50),
		"enabled":     float64(1),
		"match_mode":  "all",
		"target_type": "group",
		"target_id":   "grp-1",
		"conditions":  `[{"type":"user_group"}]`,
		"actions":     `[{"type":"bandwidth_limit"}]`,
		"created_at":  "2024-01-01",
		"updated_at":  "2024-02-01",
	}

	var model PolicyResourceModel
	mapPolicyResponse(data, &model)

	assertString(t, "ID", model.ID, "pol-1")
	assertString(t, "Name", model.Name, "BandwidthLimit")
	assertInt64(t, "Priority", model.Priority, 50)
	assertBool(t, "Enabled", model.Enabled, true)
	assertString(t, "MatchMode", model.MatchMode, "all")
	assertString(t, "TargetType", model.TargetType, "group")
	assertString(t, "TargetID", model.TargetID, "grp-1")
	assertString(t, "CreatedAt", model.CreatedAt, "2024-01-01")
	assertString(t, "UpdatedAt", model.UpdatedAt, "2024-02-01")
}

func TestMapPolicyResponse_EnabledFalse(t *testing.T) {
	data := map[string]interface{}{"id": "p1", "name": "x", "enabled": float64(0)}
	var model PolicyResourceModel
	mapPolicyResponse(data, &model)
	assertBool(t, "Enabled", model.Enabled, false)
}

// --- Auth Provider resource mapping tests ---

func TestMapAuthProviderResponse(t *testing.T) {
	data := map[string]interface{}{
		"id":                "ap-1",
		"name":              "LDAP Corp",
		"type":              "ldap",
		"captive_portal_id": "cp-1",
		"group_id":          "grp-1",
		"status":            "enabled",
		"configuration":     `{"basedn":"dc=example,dc=com"}`,
	}

	var model AuthProviderResourceModel
	mapAuthProviderResponse(data, &model)

	assertString(t, "ID", model.ID, "ap-1")
	assertString(t, "Name", model.Name, "LDAP Corp")
	assertString(t, "Type", model.Type, "ldap")
	assertString(t, "CaptivePortalID", model.CaptivePortalID, "cp-1")
	assertString(t, "GroupID", model.GroupID, "grp-1")
	assertString(t, "Status", model.Status, "enabled")
}

// --- Captive Portal resource mapping tests ---

func TestMapCaptivePortalResponse(t *testing.T) {
	data := map[string]interface{}{
		"id":                 "cp-1",
		"name":               "Guest Portal",
		"description":        "Guest WiFi portal",
		"vendor":             "meraki",
		"network_id":         "net-1",
		"splash_page":        "https://splash.example.com",
		"success_page":       "https://success.example.com",
		"portal_theme":       "modern",
		"mac_authentication": float64(1),
		"cloud_cdn":          float64(0),
		"webhook_url":        "https://hook.example.com",
	}

	var model CaptivePortalResourceModel
	mapCaptivePortalResponse(data, &model)

	assertString(t, "ID", model.ID, "cp-1")
	assertString(t, "Name", model.Name, "Guest Portal")
	assertString(t, "Vendor", model.Vendor, "meraki")
	assertString(t, "NetworkID", model.NetworkID, "net-1")
	assertString(t, "SplashPage", model.SplashPage, "https://splash.example.com")
	assertBool(t, "MacAuthentication", model.MacAuthentication, true)
	assertBool(t, "CloudCDN", model.CloudCDN, false)
	assertString(t, "WebhookURL", model.WebhookURL, "https://hook.example.com")
}

// --- Device resource mapping tests ---

func TestMapDeviceResponse(t *testing.T) {
	data := map[string]interface{}{
		"id":           "dev-1",
		"username":     "AA:BB:CC:DD:EE:FF",
		"email":        "admin@example.com",
		"firstname":    "Server",
		"lastname":     "Room",
		"notes":        "IoT device",
		"mobilephone":  "",
		"authsource":   "local",
		"orgunit":      "",
		"status":       "active",
		"creationdate": "2024-03-01",
	}

	var model DeviceResourceModel
	mapDeviceResponse(data, &model)

	assertString(t, "ID", model.ID, "dev-1")
	assertString(t, "Name", model.Name, "AA:BB:CC:DD:EE:FF")
	assertString(t, "Email", model.Email, "admin@example.com")
	assertString(t, "Firstname", model.Firstname, "Server")
	assertString(t, "Lastname", model.Lastname, "Room")
}

// --- Certificate resource mapping tests ---

func TestMapCertificateResponse(t *testing.T) {
	data := map[string]interface{}{
		"id":             "cert-1",
		"user_id":        "user-1",
		"serial":         "ABCD1234",
		"status":         "valid",
		"cn":             "john.doe",
		"subject":        "CN=john.doe,O=Example",
		"validity":       float64(365),
		"distribution":   "email",
		"hash":           "sha2",
		"expirationdate": "2025-01-01",
		"revocationdate": "",
		"creationdate":   "2024-01-01",
	}

	var model CertificateResourceModel
	mapCertificateResponse(data, &model)

	assertString(t, "ID", model.ID, "cert-1")
	assertString(t, "UserID", model.UserID, "user-1")
	assertString(t, "Serial", model.Serial, "ABCD1234")
	assertString(t, "Status", model.Status, "valid")
	assertInt64(t, "Validity", model.Validity, 365)
	assertString(t, "Hash", model.Hash, "sha2")
}

// --- Connector resource mapping tests ---

func TestMapConnectorResponse(t *testing.T) {
	data := map[string]interface{}{
		"id":            "conn-1",
		"name":          "Corp LDAP",
		"dbtype":        "ldap",
		"domain":        "example.com",
		"group":         "grp-1",
		"groupname":     "Employees",
		"status":        "enabled",
		"authsource":    "ldap",
		"basedn":        "dc=example,dc=com",
		"bind":          "cn=admin,dc=example,dc=com",
		"sync_interval": float64(60),
		"user_takeover": float64(1),
		"client_id":     "client123",
		"creationdate":  "2024-01-01",
	}

	var model ConnectorResourceModel
	mapConnectorResponse(data, &model)

	assertString(t, "ID", model.ID, "conn-1")
	assertString(t, "Name", model.Name, "Corp LDAP")
	assertString(t, "Type", model.Type, "ldap")
	assertString(t, "Domain", model.Domain, "example.com")
	assertString(t, "BaseDN", model.BaseDN, "dc=example,dc=com")
	assertBool(t, "UserTakeover", model.UserTakeover, true)
}

func TestMapConnectorResponse_PasswordNotReturned(t *testing.T) {
	data := map[string]interface{}{"id": "c1", "name": "x", "dbtype": "ldap"}
	var model ConnectorResourceModel
	mapConnectorResponse(data, &model)
	// Password and ClientSecret are write-only; should remain null
	if !model.Password.IsNull() {
		t.Error("Password should remain null")
	}
	if !model.ClientSecret.IsNull() {
		t.Error("ClientSecret should remain null")
	}
}

// --- Voucher resource mapping tests ---

func TestMapVoucherResponse(t *testing.T) {
	data := map[string]interface{}{
		"id":                 "v-1",
		"template_name":      "GuestPass",
		"voucher_format":     "alphanumeric",
		"voucher_length":     float64(8),
		"voucher_quantity":   float64(100),
		"group_id":           "grp-1",
		"orgunitid":          "ou-1",
		"voucher_deletedate": "2025-12-31",
		"voucher_devices":    float64(3),
		"voucher_duration":   "24h",
	}

	var model VoucherResourceModel
	mapVoucherResponse(data, &model)

	assertString(t, "ID", model.ID, "v-1")
	assertString(t, "TemplateName", model.TemplateName, "GuestPass")
	assertString(t, "VoucherFormat", model.VoucherFormat, "alphanumeric")
	assertString(t, "VoucherDuration", model.VoucherDuration, "24h")
}

// --- OrgUnit resource mapping tests ---

func TestMapOrgUnitResponse(t *testing.T) {
	data := map[string]interface{}{
		"id":          "ou-1",
		"name":        "Engineering",
		"description": "Eng dept",
		"parent_id":   "ou-root",
	}

	var model OrgUnitResourceModel
	mapOrgUnitResponse(data, &model)

	assertString(t, "ID", model.ID, "ou-1")
	assertString(t, "Name", model.Name, "Engineering")
	assertString(t, "Description", model.Description, "Eng dept")
	assertString(t, "ParentID", model.ParentID, "ou-root")
}

func TestMapOrgUnitResponse_NoParent(t *testing.T) {
	data := map[string]interface{}{"id": "ou-1", "name": "Root"}
	var model OrgUnitResourceModel
	mapOrgUnitResponse(data, &model)
	assertString(t, "ParentID", model.ParentID, "") // missing = empty string
}

// --- Profile resource mapping tests ---

func TestMapProfileResponse(t *testing.T) {
	data := map[string]interface{}{
		"id":            "prof-1",
		"name":          "EAP-TLS Profile",
		"description":   "Enterprise profile",
		"type":          "EAP-TLS",
		"configuration": `{"cert":"pem-data"}`,
	}

	var model ProfileResourceModel
	mapProfileResponse(data, &model)

	assertString(t, "ID", model.ID, "prof-1")
	assertString(t, "Name", model.Name, "EAP-TLS Profile")
	assertString(t, "Description", model.Description, "Enterprise profile")
	assertString(t, "Type", model.Type, "EAP-TLS")
}

// --- Edge case tests for helpers ---

func TestBoolFromIntAPI_StringTrue(t *testing.T) {
	data := map[string]interface{}{"k": "true"}
	v := boolFromIntAPI(data, "k")
	if !v.ValueBool() {
		t.Errorf("expected true for string 'true'")
	}
}

func TestStringFromAPI_NilValue(t *testing.T) {
	data := map[string]interface{}{"k": nil}
	v := stringFromAPI(data, "k")
	if v.ValueString() != "" {
		t.Errorf("expected empty string for nil, got %q", v.ValueString())
	}
}

func TestIntFromAPI_Int64Type(t *testing.T) {
	data := map[string]interface{}{"k": int64(42)}
	v := intFromAPI(data, "k")
	if v.ValueInt64() != 42 {
		t.Errorf("expected 42, got %d", v.ValueInt64())
	}
}

func TestIntFromAPINullable_Missing(t *testing.T) {
	data := map[string]interface{}{}
	v := intFromAPINullable(data, "k")
	if !v.IsNull() {
		t.Errorf("expected null for missing key")
	}
}

func TestStringFromAPINullable(t *testing.T) {
	data := map[string]interface{}{"k": "val"}
	v := stringFromAPINullable(data, "k")
	if v.ValueString() != "val" {
		t.Errorf("expected 'val', got %q", v.ValueString())
	}

	v2 := stringFromAPINullable(data, "missing")
	if !v2.IsNull() {
		t.Errorf("expected null for missing key")
	}
}
