package utility

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestValidateFieldType tests field type validation
func TestValidateFieldType(t *testing.T) {
	tests := []struct {
		name         string
		value        interface{}
		expectedType string
		want         bool
	}{
		{"string valid", "test", "string", true},
		{"string invalid", 123, "string", false},
		{"int valid from float64", float64(123), "int", true},
		{"int valid from int", int(123), "int", true},
		{"int invalid", "123", "int", false},
		{"number valid float64", float64(123.45), "number", true},
		{"number valid int", int(123), "number", true},
		{"number invalid", "123.45", "number", false},
		{"bool valid true", true, "bool", true},
		{"bool valid false", false, "bool", true},
		{"bool invalid", "true", "bool", false},
		{"null valid", nil, "null", true},
		{"null invalid", "null", "null", false},
		{"unknown type", "value", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateFieldType(tt.value, tt.expectedType)
			if got != tt.want {
				t.Errorf("ValidateFieldType(%v, %v) = %v, want %v", tt.value, tt.expectedType, got, tt.want)
			}
		})
	}
}

// TestCheckPatchPayload tests PATCH payload validation
func TestCheckPatchPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	allowedFields := FieldSpecMap{
		"Name":    {Name: "Name", Expected: "string"},
		"Age":     {Name: "Age", Expected: "number"},
		"Active":  {Name: "Active", Expected: "bool"},
	}

	tests := []struct {
		name       string
		payload    string
		wantValid  bool
		wantStatus int
	}{
		{
			name:       "valid payload",
			payload:    `{"Name":"test","Age":25,"Active":true}`,
			wantValid:  true,
			wantStatus: 0,
		},
		{
			name:       "valid partial payload",
			payload:    `{"Name":"test"}`,
			wantValid:  true,
			wantStatus: 0,
		},
		{
			name:       "invalid JSON",
			payload:    `{"Name":"test"`,
			wantValid:  false,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid field",
			payload:    `{"Name":"test","InvalidField":"value"}`,
			wantValid:  false,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid type",
			payload:    `{"Name":123}`,
			wantValid:  false,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid age type",
			payload:    `{"Age":"25"}`,
			wantValid:  false,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid bool type",
			payload:    `{"Active":"true"}`,
			wantValid:  false,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/test", strings.NewReader(tt.payload))
			c.Request.Header.Set("Content-Type", "application/json")

			patchData, valid := CheckPatchPayload(c, allowedFields)

			if valid != tt.wantValid {
				t.Errorf("CheckPatchPayload() valid = %v, want %v", valid, tt.wantValid)
			}

			if !tt.wantValid && w.Code != tt.wantStatus {
				t.Errorf("CheckPatchPayload() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantValid && patchData == nil {
				t.Errorf("CheckPatchPayload() returned nil patchData for valid payload")
			}
		})
	}
}

// TestBindMapToStruct tests map to struct binding
func TestBindMapToStruct(t *testing.T) {
	type TestStruct struct {
		Name   string `json:"Name"`
		Age    int    `json:"Age"`
		Active bool   `json:"Active"`
	}

	tests := []struct {
		name    string
		data    map[string]interface{}
		wantErr bool
		want    TestStruct
	}{
		{
			name: "valid binding",
			data: map[string]interface{}{
				"Name":   "John",
				"Age":    float64(30), // JSON numbers are float64
				"Active": true,
			},
			wantErr: false,
			want:    TestStruct{Name: "John", Age: 30, Active: true},
		},
		{
			name: "partial binding",
			data: map[string]interface{}{
				"Name": "Jane",
			},
			wantErr: false,
			want:    TestStruct{Name: "Jane", Age: 0, Active: false},
		},
		{
			name:    "empty map",
			data:    map[string]interface{}{},
			wantErr: false,
			want:    TestStruct{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result TestStruct
			err := BindMapToStruct(tt.data, &result)

			if (err != nil) != tt.wantErr {
				t.Errorf("BindMapToStruct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result.Name != tt.want.Name {
					t.Errorf("BindMapToStruct() Name = %v, want %v", result.Name, tt.want.Name)
				}
				if result.Age != tt.want.Age {
					t.Errorf("BindMapToStruct() Age = %v, want %v", result.Age, tt.want.Age)
				}
				if result.Active != tt.want.Active {
					t.Errorf("BindMapToStruct() Active = %v, want %v", result.Active, tt.want.Active)
				}
			}
		})
	}
}

// TestCheckAndBindPatchPayload tests the combined validation and binding
func TestCheckAndBindPatchPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type UpdateStruct struct {
		Name   *string `json:"Name,omitempty"`
		Age    *int    `json:"Age,omitempty"`
		Active *bool   `json:"Active,omitempty"`
	}

	allowedFields := FieldSpecMap{
		"Name":   {Name: "Name", Expected: "string"},
		"Age":    {Name: "Age", Expected: "number"},
		"Active": {Name: "Active", Expected: "bool"},
	}

	tests := []struct {
		name       string
		payload    string
		wantValid  bool
		wantStatus int
	}{
		{
			name:       "valid payload",
			payload:    `{"Name":"test","Age":25}`,
			wantValid:  true,
			wantStatus: 0,
		},
		{
			name:       "invalid field",
			payload:    `{"Name":"test","BadField":"value"}`,
			wantValid:  false,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid type",
			payload:    `{"Name":123}`,
			wantValid:  false,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/test", strings.NewReader(tt.payload))
			c.Request.Header.Set("Content-Type", "application/json")

			var target UpdateStruct
			valid := CheckAndBindPatchPayload(c, allowedFields, &target)

			if valid != tt.wantValid {
				t.Errorf("CheckAndBindPatchPayload() valid = %v, want %v", valid, tt.wantValid)
			}

			if !tt.wantValid && w.Code != tt.wantStatus {
				t.Errorf("CheckAndBindPatchPayload() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

// TestRedfishError tests Redfish error response generation
func TestRedfishError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		statusCode int
		message    string
		messageID  string
	}{
		{
			name:       "not found error",
			statusCode: http.StatusNotFound,
			message:    "Resource not found",
			messageID:  "ResourceNotFound",
		},
		{
			name:       "bad request error",
			statusCode: http.StatusBadRequest,
			message:    "Invalid property",
			messageID:  "InvalidProperty",
		},
		{
			name:       "internal error",
			statusCode: http.StatusInternalServerError,
			message:    "Internal server error",
			messageID:  "InternalError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			RedfishError(c, tt.statusCode, tt.message, tt.messageID)

			if w.Code != tt.statusCode {
				t.Errorf("RedfishError() status = %v, want %v", w.Code, tt.statusCode)
			}

			// Verify response contains error structure
			body := w.Body.String()
			if !strings.Contains(body, "error") {
				t.Errorf("RedfishError() response missing 'error' field")
			}
			if !strings.Contains(body, tt.message) {
				t.Errorf("RedfishError() response missing message: %v", tt.message)
			}
		})
	}
}

// TestFieldSpec tests FieldSpec structure
func TestFieldSpec(t *testing.T) {
	spec := FieldSpec{
		Name:     "TestField",
		Expected: "string",
	}

	if spec.Name != "TestField" {
		t.Errorf("FieldSpec.Name = %v, want TestField", spec.Name)
	}
	if spec.Expected != "string" {
		t.Errorf("FieldSpec.Expected = %v, want string", spec.Expected)
	}
}

// TestFieldSpecMap tests FieldSpecMap usage
func TestFieldSpecMap(t *testing.T) {
	specMap := FieldSpecMap{
		"Field1": {Name: "Field1", Expected: "string"},
		"Field2": {Name: "Field2", Expected: "number"},
	}

	if len(specMap) != 2 {
		t.Errorf("FieldSpecMap length = %v, want 2", len(specMap))
	}

	if spec, exists := specMap["Field1"]; !exists {
		t.Error("FieldSpecMap missing Field1")
	} else if spec.Expected != "string" {
		t.Errorf("FieldSpecMap[Field1].Expected = %v, want string", spec.Expected)
	}
}
