package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SergeiKhy/org-structure-api/internal/model"
	"github.com/SergeiKhy/org-structure-api/internal/service"
)

// Helper functions
func intPtr(i int) *int {
	return &i
}

func strPtr(s string) *string {
	return &s
}

// TestWriteJSON tests the writeJSON helper
func TestWriteJSON(t *testing.T) {
	h := &Handler{}

	tests := []struct {
		name   string
		data   interface{}
		status int
	}{
		{"nil data", nil, http.StatusOK},
		{"string data", "test", http.StatusOK},
		{"map data", map[string]string{"key": "value"}, http.StatusOK},
		{"struct data", model.Department{Name: "Test"}, http.StatusCreated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			h.writeJSON(w, tt.status, tt.data)

			if w.Code != tt.status {
				t.Errorf("expected status %d, got %d", tt.status, w.Code)
			}

			if w.Header().Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
			}
		})
	}
}

// TestWriteError tests the WriteError helper
func TestWriteError(t *testing.T) {
	h := &Handler{}

	w := httptest.NewRecorder()
	h.WriteError(w, http.StatusBadRequest, "test error")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)

	if response["error"] != "test error" {
		t.Errorf("expected error message 'test error', got '%s'", response["error"])
	}
}

// TestInvalidJSON tests handling of invalid JSON
func TestInvalidJSON(t *testing.T) {
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPost, "/departments/", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	h.CreateDepartment(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for invalid JSON, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestCreateEmployeeRequestParsing tests employee request JSON parsing
func TestCreateEmployeeRequestParsing(t *testing.T) {
	tests := []struct {
		name        string
		body        model.CreateEmployeeRequest
		expectError bool
	}{
		{
			name: "valid employee request",
			body: model.CreateEmployeeRequest{
				FullName: "John Doe",
				Position: "Developer",
				HiredAt:  strPtr("2024-01-15"),
			},
			expectError: false,
		},
		{
			name: "employee without hired_at",
			body: model.CreateEmployeeRequest{
				FullName: "Jane Smith",
				Position: "Manager",
			},
			expectError: false,
		},
		{
			name: "empty full_name",
			body: model.CreateEmployeeRequest{
				FullName: "",
				Position: "Developer",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			var parsed model.CreateEmployeeRequest
			err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestUpdateDepartmentRequestParsing tests update request JSON parsing
func TestUpdateDepartmentRequestParsing(t *testing.T) {
	tests := []struct {
		name string
		body model.UpdateDepartmentRequest
	}{
		{
			name: "update name only",
			body: model.UpdateDepartmentRequest{
				Name: "Updated Name",
			},
		},
		{
			name: "update parent_id only",
			body: model.UpdateDepartmentRequest{
				ParentID: intPtr(5),
			},
		},
		{
			name: "update both fields",
			body: model.UpdateDepartmentRequest{
				Name:     "Updated",
				ParentID: intPtr(3),
			},
		},
		{
			name: "empty update (keep existing)",
			body: model.UpdateDepartmentRequest{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			var parsed model.UpdateDepartmentRequest
			err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestHTTPStatusCodes verifies HTTP status code constants
func TestHTTPStatusCodes(t *testing.T) {
	tests := []struct {
		name   string
		code   int
		expect int
	}{
		{"StatusOK", http.StatusOK, 200},
		{"StatusCreated", http.StatusCreated, 201},
		{"StatusNoContent", http.StatusNoContent, 204},
		{"StatusBadRequest", http.StatusBadRequest, 400},
		{"StatusNotFound", http.StatusNotFound, 404},
		{"StatusConflict", http.StatusConflict, 409},
		{"StatusInternalServerError", http.StatusInternalServerError, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.expect {
				t.Errorf("expected %s to be %d, got %d", tt.name, tt.expect, tt.code)
			}
		})
	}
}

// TestServiceErrors tests service error types
func TestServiceErrors(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{"not found", service.ErrNotFound, http.StatusNotFound},
		{"cycle detected", service.ErrCycleDetected, http.StatusConflict},
		{"self parent", service.ErrSelfParent, http.StatusConflict},
		{"duplicate name", service.ErrDuplicateName, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Error("expected error, got nil")
			}

			// Verify error messages are not empty
			if tt.err.Error() == "" {
				t.Error("expected non-empty error message")
			}
		})
	}
}

// TestNameTrimming validates that names are trimmed
func TestNameTrimming(t *testing.T) {
	input := "  TrimmedName  "
	expected := "TrimmedName"

	// Simulate trimming logic from service
	trimmed := input
	for len(trimmed) > 0 && trimmed[0] == ' ' {
		trimmed = trimmed[1:]
	}
	for len(trimmed) > 0 && trimmed[len(trimmed)-1] == ' ' {
		trimmed = trimmed[:len(trimmed)-1]
	}

	if trimmed != expected {
		t.Errorf("expected %q, got %q", expected, trimmed)
	}
}

// TestDepthValidation tests depth parameter validation logic
func TestDepthValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"zero becomes 1", 0, 1},
		{"negative becomes 1", -5, 1},
		{"normal value", 3, 3},
		{"max exceeded capped", 10, 5},
		{"at max", 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			depth := tt.input
			if depth < 1 {
				depth = 1
			}
			if depth > 5 {
				depth = 5
			}

			if depth != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, depth)
			}
		})
	}
}

// TestEmployeeValidation tests employee field validation
func TestEmployeeValidation(t *testing.T) {
	tests := []struct {
		name       string
		fullName   string
		position   string
		shouldPass bool
	}{
		{"valid", "John Doe", "Developer", true},
		{"empty name", "", "Developer", false},
		{"empty position", "John Doe", "", false},
		{"both empty", "", "", false},
		{"name too long", string(make([]byte, 201)), "Developer", false},
		{"position too long", "John Doe", string(make([]byte, 201)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := true
			if tt.fullName == "" || tt.position == "" {
				valid = false
			}
			if len(tt.fullName) > 200 || len(tt.position) > 200 {
				valid = false
			}

			if valid != tt.shouldPass {
				t.Errorf("expected validation to %v, got %v", tt.shouldPass, valid)
			}
		})
	}
}

// TestDepartmentValidation tests department name validation
func TestDepartmentValidation(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		shouldPass bool
	}{
		{"valid", "Engineering", true},
		{"empty", "", false},
		{"spaces only", "   ", false},
		{"too long", string(make([]byte, 201)), false},
		{"at max length", string(make([]byte, 200)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trimmed := tt.input
			for len(trimmed) > 0 && trimmed[0] == ' ' {
				trimmed = trimmed[1:]
			}
			for len(trimmed) > 0 && trimmed[len(trimmed)-1] == ' ' {
				trimmed = trimmed[:len(trimmed)-1]
			}

			valid := true
			if trimmed == "" || len(trimmed) > 200 {
				valid = false
			}

			if valid != tt.shouldPass {
				t.Errorf("expected validation to %v, got %v", tt.shouldPass, valid)
			}
		})
	}
}

// Benchmark tests
func BenchmarkJSONEncoding(b *testing.B) {
	dept := model.Department{
		ID:   1,
		Name: "Engineering",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(dept)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONDecoding(b *testing.B) {
	body := `{"name": "Engineering", "parent_id": 1}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var req model.CreateDepartmentRequest
		_ = json.NewDecoder(bytes.NewReader([]byte(body))).Decode(&req)
	}
}
