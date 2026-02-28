package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew_RequiresEndpoint(t *testing.T) {
	_, err := New(&Config{CompanyID: "test", APIToken: "tok"})
	if err == nil {
		t.Fatal("expected error for missing endpoint")
	}
}

func TestNew_RequiresCompanyID(t *testing.T) {
	_, err := New(&Config{APIEndpoint: "http://localhost", APIToken: "tok"})
	if err == nil {
		t.Fatal("expected error for missing company_id")
	}
}

func TestNew_RequiresAuth(t *testing.T) {
	_, err := New(&Config{APIEndpoint: "http://localhost", CompanyID: "test"})
	if err == nil {
		t.Fatal("expected error for missing auth")
	}
}

func TestNew_AcceptsAPIToken(t *testing.T) {
	c, err := New(&Config{APIEndpoint: "http://localhost", CompanyID: "test", APIToken: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestClient_Read(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/company123/networks/net-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/hal+json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "net-1",
			"nasname": "TestNetwork",
		})
	}))
	defer server.Close()

	c, _ := New(&Config{
		APIEndpoint: server.URL,
		CompanyID:   "company123",
		APIToken:    "test-token",
	})

	result, err := c.Read("networks", "net-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["id"] != "net-1" {
		t.Errorf("expected id 'net-1', got %v", result["id"])
	}
}

func TestClient_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["nasname"] != "NewNetwork" {
			t.Errorf("unexpected body: %v", body)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "new-id",
			"nasname": "NewNetwork",
		})
	}))
	defer server.Close()

	c, _ := New(&Config{
		APIEndpoint: server.URL,
		CompanyID:   "company123",
		APIToken:    "test-token",
	})

	result, err := c.Create("networks", map[string]interface{}{"nasname": "NewNetwork"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["id"] != "new-id" {
		t.Errorf("expected id 'new-id', got %v", result["id"])
	}
}

func TestClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c, _ := New(&Config{
		APIEndpoint: server.URL,
		CompanyID:   "company123",
		APIToken:    "test-token",
	})

	err := c.Delete("networks", "net-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/hal+json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"_embedded": map[string]interface{}{
				"networks": []map[string]interface{}{
					{"id": "n1", "nasname": "Net1"},
					{"id": "n2", "nasname": "Net2"},
				},
			},
			"total_items": 2,
			"page_count":  1,
		})
	}))
	defer server.Close()

	c, _ := New(&Config{
		APIEndpoint: server.URL,
		CompanyID:   "company123",
		APIToken:    "test-token",
	})

	items, err := c.List("networks", "networks")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestClient_404_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := New(&Config{
		APIEndpoint: server.URL,
		CompanyID:   "company123",
		APIToken:    "test-token",
	})

	_, err := c.Read("networks", "nonexistent")
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %v", err)
	}
}

func TestClient_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"type":   "http://www.w3.org/Protocols/rfc2616/rfc2616-sec10.html",
			"title":  "Unprocessable Entity",
			"status": 422,
			"detail": "Validation failed",
		})
	}))
	defer server.Close()

	c, _ := New(&Config{
		APIEndpoint: server.URL,
		CompanyID:   "company123",
		APIToken:    "test-token",
	})

	_, err := c.Create("networks", map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Status != 422 {
		t.Errorf("expected status 422, got %d", apiErr.Status)
	}
}
