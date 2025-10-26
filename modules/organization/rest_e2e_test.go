package main_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

const RestBaseURL = "http://localhost:8082"
const GRPCAddress = "localhost:9082"
const MockJWTToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI3NzBlODQwMC1lMjliLTQxZDQtYTcxNi00NDY2NTU0NDAwMDAifQ.fake"

// CreateOrgRequest represents the REST API create request
type CreateOrgRequest struct {
	Name      string `json:"name"`
	Subdomain string `json:"subdomain"`
}

// OrganizationResponse represents the REST API response
type OrganizationResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Subdomain   string `json:"subdomain"`
	SchemaName  string `json:"schema_name"`
	OwnerUserID string `json:"owner_user_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details"`
}

func makeRESTRequest(t *testing.T, method, endpoint string, body interface{}) (*http.Response, []byte) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, RestBaseURL+endpoint, reqBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", MockJWTToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	return resp, respBody
}

// TestCreateOrganizationRESTAPI tests REST API create endpoint
func TestCreateOrganizationRESTAPI(t *testing.T) {
	req := CreateOrgRequest{
		Name:      "REST E2E Test Org",
		Subdomain: "rest-e2e-" + uuid.New().String()[:8],
	}

	resp, respBody := makeRESTRequest(t, "POST", "/api/organizations", req)

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
		t.Logf("Response: %s", string(respBody))
	}

	var orgResp OrganizationResponse
	if err := json.Unmarshal(respBody, &orgResp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if orgResp.ID == "" {
		t.Error("Organization ID is empty")
	}
	if orgResp.Name != req.Name {
		t.Errorf("Expected name %q, got %q", req.Name, orgResp.Name)
	}
	if orgResp.Subdomain != req.Subdomain {
		t.Errorf("Expected subdomain %q, got %q", req.Subdomain, orgResp.Subdomain)
	}

	t.Logf("✓ Created organization via REST: ID=%s, Name=%s", orgResp.ID, orgResp.Name)
}

// TestGetOrganizationRESTAPI tests REST API get endpoint
func TestGetOrganizationRESTAPI(t *testing.T) {
	// Create first
	createReq := CreateOrgRequest{
		Name:      "Get REST E2E Test",
		Subdomain: "get-rest-" + uuid.New().String()[:8],
	}

	resp1, respBody1 := makeRESTRequest(t, "POST", "/api/organizations", createReq)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create organization: %d", resp1.StatusCode)
	}

	var createResp OrganizationResponse
	if err := json.Unmarshal(respBody1, &createResp); err != nil {
		t.Fatalf("Failed to unmarshal create response: %v", err)
	}
	t.Logf("Created organization: %s", createResp.ID)

	// Then get
	resp2, respBody2 := makeRESTRequest(t, "GET", "/api/organizations/"+createResp.ID, nil)

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp2.StatusCode)
		t.Logf("Response: %s", string(respBody2))
		return
	}

	var getResp OrganizationResponse
	if err := json.Unmarshal(respBody2, &getResp); err != nil {
		t.Fatalf("Failed to unmarshal get response: %v", err)
	}

	if getResp.ID != createResp.ID {
		t.Errorf("Expected ID %q, got %q", createResp.ID, getResp.ID)
	}
	if getResp.OwnerUserID == "" {
		t.Error("Owner user ID is empty")
	}

	t.Logf("✓ Retrieved organization via REST: ID=%s, Name=%s, Owner=%s", getResp.ID, getResp.Name, getResp.OwnerUserID)
}

// TestHealthCheckRESTAPI tests health endpoint
func TestHealthCheckRESTAPI(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(RestBaseURL + "/health")
	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
		return
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	var healthResp map[string]string
	if err := json.Unmarshal(respBody, &healthResp); err != nil {
		t.Fatalf("Failed to unmarshal health response: %v", err)
	}

	status, ok := healthResp["status"]
	if !ok {
		t.Error("Status field is missing from health response")
	}
	if status != "ok" && status != "healthy" {
		t.Errorf("Expected status 'ok' or 'healthy', got %q", status)
	}

	t.Logf("✓ Health check passed: %s", status)
}

// TestCreateWithDuplicateSubdomainRESTAPI tests duplicate subdomain error
func TestCreateWithDuplicateSubdomainRESTAPI(t *testing.T) {
	subdomain := "dup-rest-" + uuid.New().String()[:8]

	// Create first
	req1 := CreateOrgRequest{
		Name:      "First REST Org",
		Subdomain: subdomain,
	}
	resp1, _ := makeRESTRequest(t, "POST", "/api/organizations", req1)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create first organization: %d", resp1.StatusCode)
	}
	t.Logf("✓ Created first organization")

	// Try to create second with same subdomain
	req2 := CreateOrgRequest{
		Name:      "Second REST Org",
		Subdomain: subdomain,
	}
	resp2, respBody2 := makeRESTRequest(t, "POST", "/api/organizations", req2)

	if resp2.StatusCode != http.StatusConflict {
		t.Errorf("Expected status 409 Conflict, got %d", resp2.StatusCode)
		t.Logf("Response: %s", string(respBody2))
	} else {
		t.Logf("✓ Got expected 409 Conflict for duplicate subdomain")
	}
}

// TestGetNonExistentOrganizationRESTAPI tests 404 error
func TestGetNonExistentOrganizationRESTAPI(t *testing.T) {
	fakeID := "550e8400-e29b-41d4-a716-446655440099"

	resp, respBody := makeRESTRequest(t, "GET", "/api/organizations/"+fakeID, nil)

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 Not Found, got %d", resp.StatusCode)
		t.Logf("Response: %s", string(respBody))
	} else {
		t.Logf("✓ Got expected 404 for non-existent organization")
	}
}

// TestInvalidJSONRESTAPI tests malformed request handling
func TestInvalidJSONRESTAPI(t *testing.T) {
	reqBody := []byte("{invalid json")

	req, err := http.NewRequest("POST", RestBaseURL+"/api/organizations", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", MockJWTToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request, got %d", resp.StatusCode)
	} else {
		t.Logf("✓ Got expected 400 for invalid JSON")
	}
}

// TestMissingAuthorizationRESTAPI tests missing JWT token
func TestMissingAuthorizationRESTAPI(t *testing.T) {
	req := CreateOrgRequest{
		Name:      "No Auth Test",
		Subdomain: "no-auth-" + uuid.New().String()[:8],
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", RestBaseURL+"/api/organizations", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	// No Authorization header

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 Unauthorized, got %d", resp.StatusCode)
	} else {
		t.Logf("✓ Got expected 401 for missing authorization")
	}
}

// TestInvalidUUIDFormatRESTAPI tests invalid UUID in path
func TestInvalidUUIDFormatRESTAPI(t *testing.T) {
	resp, respBody := makeRESTRequest(t, "GET", "/api/organizations/not-a-uuid", nil)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request, got %d", resp.StatusCode)
		t.Logf("Response: %s", string(respBody))
	} else {
		t.Logf("✓ Got expected 400 for invalid UUID format")
	}
}
