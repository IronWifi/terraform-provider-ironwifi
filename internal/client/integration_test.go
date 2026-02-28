// +build integration

package client

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// These tests run against a live IronWiFi API instance.
// Run with: go test -tags integration ./internal/client/ -v
//
// Required env vars:
//   IRONWIFI_TEST_ENDPOINT (default: http://localhost:8080)
//   IRONWIFI_TEST_USERNAME
//   IRONWIFI_TEST_PASSWORD
//   IRONWIFI_TEST_COMPANY_ID
//   IRONWIFI_TEST_CLIENT_ID (default: testclient)
//   IRONWIFI_TEST_CLIENT_SECRET (default: testpass)

func getTestClient(t *testing.T) *Client {
	t.Helper()

	endpoint := os.Getenv("IRONWIFI_TEST_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:8080"
	}
	username := os.Getenv("IRONWIFI_TEST_USERNAME")
	password := os.Getenv("IRONWIFI_TEST_PASSWORD")
	companyID := os.Getenv("IRONWIFI_TEST_COMPANY_ID")
	clientID := os.Getenv("IRONWIFI_TEST_CLIENT_ID")
	clientSecret := os.Getenv("IRONWIFI_TEST_CLIENT_SECRET")

	if username == "" || password == "" || companyID == "" {
		t.Skip("IRONWIFI_TEST_USERNAME, IRONWIFI_TEST_PASSWORD, and IRONWIFI_TEST_COMPANY_ID must be set")
	}
	if clientID == "" {
		clientID = "testclient"
	}
	if clientSecret == "" {
		clientSecret = "testpass"
	}

	c, err := New(&Config{
		APIEndpoint:  endpoint,
		Username:     username,
		Password:     password,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		CompanyID:    companyID,
		UserAgent:    "terraform-provider-ironwifi-test",
	})
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}
	return c
}

func TestIntegration_GroupCRUD(t *testing.T) {
	c := getTestClient(t)
	uniqueName := fmt.Sprintf("tf-test-group-%d", time.Now().UnixMilli())

	// CREATE
	created, err := c.Create("groups", map[string]interface{}{
		"groupname":   uniqueName,
		"description": "Terraform integration test",
	})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	id, ok := created["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected non-empty id, got: %v", created["id"])
	}
	t.Logf("Created group: %s (id=%s)", uniqueName, id)

	// READ
	fetched, err := c.Read("groups", id)
	if err != nil {
		t.Fatalf("read group: %v", err)
	}
	if fmt.Sprintf("%v", fetched["groupname"]) != uniqueName {
		t.Errorf("expected groupname=%q, got %v", uniqueName, fetched["groupname"])
	}

	// UPDATE
	updatedName := uniqueName + "-updated"
	updated, err := c.Update("groups", id, map[string]interface{}{
		"groupname":   updatedName,
		"description": "Updated by Terraform test",
	})
	if err != nil {
		t.Fatalf("update group: %v", err)
	}
	if fmt.Sprintf("%v", updated["groupname"]) != updatedName {
		t.Errorf("expected updated name=%q, got %v", updatedName, updated["groupname"])
	}

	// LIST
	items, err := c.List("groups", "groups")
	if err != nil {
		t.Fatalf("list groups: %v", err)
	}
	found := false
	for _, item := range items {
		if fmt.Sprintf("%v", item["id"]) == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created group not found in list")
	}

	// DELETE
	err = c.Delete("groups", id)
	if err != nil {
		t.Fatalf("delete group: %v", err)
	}

	// Verify deleted
	_, err = c.Read("groups", id)
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError after delete, got: %v", err)
	}
	t.Logf("Group CRUD lifecycle complete")
}

func TestIntegration_OrgUnitCRUD(t *testing.T) {
	c := getTestClient(t)
	uniqueName := fmt.Sprintf("tf-test-ou-%d", time.Now().UnixMilli())

	// CREATE
	created, err := c.Create("orgunits", map[string]interface{}{
		"name":        uniqueName,
		"description": "Terraform integration test",
	})
	if err != nil {
		t.Fatalf("create org unit: %v", err)
	}
	id, ok := created["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected non-empty id, got: %v", created["id"])
	}
	t.Logf("Created org unit: %s (id=%s)", uniqueName, id)

	// READ
	fetched, err := c.Read("orgunits", id)
	if err != nil {
		t.Fatalf("read org unit: %v", err)
	}
	if fmt.Sprintf("%v", fetched["name"]) != uniqueName {
		t.Errorf("expected name=%q, got %v", uniqueName, fetched["name"])
	}

	// UPDATE
	updatedName := uniqueName + "-updated"
	updated, err := c.Update("orgunits", id, map[string]interface{}{
		"name": updatedName,
	})
	if err != nil {
		t.Fatalf("update org unit: %v", err)
	}
	if fmt.Sprintf("%v", updated["name"]) != updatedName {
		t.Errorf("expected updated name=%q, got %v", updatedName, updated["name"])
	}

	// DELETE
	err = c.Delete("orgunits", id)
	if err != nil {
		t.Fatalf("delete org unit: %v", err)
	}

	// Verify deleted
	_, err = c.Read("orgunits", id)
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError after delete, got: %v", err)
	}
	t.Logf("Org unit CRUD lifecycle complete")
}

func TestIntegration_UserCRUD(t *testing.T) {
	c := getTestClient(t)
	uniqueUsername := fmt.Sprintf("tf-test-%d@example.com", time.Now().UnixMilli())

	// CREATE
	created, err := c.Create("users", map[string]interface{}{
		"username":   uniqueUsername,
		"password":   "TestP@ss12345",
		"email":      uniqueUsername,
		"firstname":  "Terraform",
		"lastname":   "Test",
		"user_type":  "e",
		"authsource": "local",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	id, ok := created["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected non-empty id, got: %v", created["id"])
	}
	t.Logf("Created user: %s (id=%s)", uniqueUsername, id)

	// READ
	fetched, err := c.Read("users", id)
	if err != nil {
		t.Fatalf("read user: %v", err)
	}
	if fmt.Sprintf("%v", fetched["username"]) != uniqueUsername {
		t.Errorf("expected username=%q, got %v", uniqueUsername, fetched["username"])
	}

	// UPDATE
	_, err = c.Update("users", id, map[string]interface{}{
		"firstname": "TerraformUpdated",
		"notes":     "Updated by integration test",
	})
	if err != nil {
		t.Fatalf("update user: %v", err)
	}

	// Verify update
	updatedUser, err := c.Read("users", id)
	if err != nil {
		t.Fatalf("re-read user: %v", err)
	}
	if fmt.Sprintf("%v", updatedUser["firstname"]) != "TerraformUpdated" {
		t.Errorf("expected firstname=TerraformUpdated, got %v", updatedUser["firstname"])
	}

	// DELETE
	err = c.Delete("users", id)
	if err != nil {
		t.Fatalf("delete user: %v", err)
	}

	// Verify deleted
	_, err = c.Read("users", id)
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError after delete, got: %v", err)
	}
	t.Logf("User CRUD lifecycle complete")
}

func TestIntegration_NetworkList(t *testing.T) {
	c := getTestClient(t)

	items, err := c.List("networks", "networks")
	if err != nil {
		t.Fatalf("list networks: %v", err)
	}
	t.Logf("Found %d networks", len(items))

	for i, item := range items {
		if i >= 3 {
			break
		}
		t.Logf("  Network: %v (id=%v, region=%v)", item["nasname"], item["id"], item["region"])
	}
}
