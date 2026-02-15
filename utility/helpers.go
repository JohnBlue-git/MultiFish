package utility

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ========== Validation Utilities ==========

// ValidateFieldType checks if a value matches the expected type
func ValidateFieldType(value interface{}, expectedType string) bool {
	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "int", "number":
		switch value.(type) {
		case float64:
			return true
		case int:
			return true
		}
		return false
	case "bool":
		_, ok := value.(bool)
		return ok
	case "null":
		return value == nil
	default:
		return false
	}
}

// ========== PATCH Payload Utilities ==========

// CheckPatchPayloadHelper validates PATCH payload fields against allowed fields and types
// This is a pure validation function that doesn't depend on gin.Context
// Returns error response if validation fails, nil otherwise
func CheckPatchPayloadHelper(patch map[string]interface{}, allowedFields FieldSpecMap) *RedfishErrorResponse {
	// Check if all provided fields are allowed and have correct types
	for field, value := range patch {
		fieldSpec, exists := allowedFields[field]
		if !exists {
			allowedFieldNames := make([]string, 0, len(allowedFields))
			for name := range allowedFields {
				allowedFieldNames = append(allowedFieldNames, name)
			}
			return &RedfishErrorResponse{
				Error: RedfishErrorDetail{
					Code:    "Base.1.0.InvalidProperty",
					Message: fmt.Sprintf("field '%s' is not allowed. Allowed fields: %v", field, allowedFieldNames),
				},
			}
		}
		
		// Type validation
		if !ValidateFieldType(value, fieldSpec.Expected) {
			return &RedfishErrorResponse{
				Error: RedfishErrorDetail{
					Code:    "Base.1.0.InvalidPropertyType",
					Message: fmt.Sprintf("field '%s' has invalid type. Expected %s, got %T", field, fieldSpec.Expected, value),
				},
			}
		}
	}
	
	return nil
}

// CheckPatchPayload parses and validates PATCH payload, sends error response if validation fails
// Returns the parsed payload map and true if successful, nil and false otherwise
func CheckPatchPayload(c *gin.Context, allowedFields FieldSpecMap) (map[string]interface{}, bool) {
	var patch map[string]interface{}
	
	// Parse the JSON body
	if err := c.ShouldBindJSON(&patch); err != nil {
		c.JSON(http.StatusBadRequest, &RedfishErrorResponse{
			Error: RedfishErrorDetail{
				Code:    "Base.1.0.InvalidJSON",
				Message: fmt.Sprintf("Invalid request body: %v", err),
			},
		})
		return nil, false
	}
	
	// Validate using the helper
	if errResp := CheckPatchPayloadHelper(patch, allowedFields); errResp != nil {
		c.JSON(http.StatusBadRequest, errResp)
		return nil, false
	}
	
	return patch, true
}

// CheckAndBindPatchPayload validates the PATCH payload and binds it to a target struct
// This is a higher-level utility that combines CheckPatchPayload with binding logic
// Returns true if successful, false if validation failed (error already sent to client)
func CheckAndBindPatchPayload(c *gin.Context, allowedFields FieldSpecMap, target interface{}) bool {
	patchData, valid := CheckPatchPayload(c, allowedFields)
	if !valid {
		return false
	}
	
	// Use reflection or a manual approach to bind patchData to target
	// For now, we'll use json encoding/decoding as a simple approach
	if err := BindMapToStruct(patchData, target); err != nil {
		RedfishError(c, http.StatusInternalServerError, 
			fmt.Sprintf("Failed to bind data: %v", err), "InternalError")
		return false
	}
	
	return true
}

// BindMapToStruct converts a map[string]interface{} to a struct
// This is a simple utility to avoid manual field-by-field conversion
func BindMapToStruct(data map[string]interface{}, target interface{}) error {
	// Use JSON marshaling as a simple conversion mechanism
	// This works because the struct tags match the map keys
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data of type %T to JSON: %w", data, err)
	}
	
	if err := json.Unmarshal(jsonBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON to target type %T: %w. Check JSON structure matches expected type", target, err)
	}
	
	return nil
}

// ========== Type Definitions ==========

// FieldSpec defines a field specification for validation
type FieldSpec struct {
	Name     string // Field name
	Expected string // Expected type: "string", "int", "bool", "number"
}

// FieldSpecMap is a map of field specifications for easy lookup
type FieldSpecMap map[string]FieldSpec

// ========== Redfish Error Response ==========

// RedfishErrorResponse represents a Redfish-compliant error response
type RedfishErrorResponse struct {
	Error RedfishErrorDetail `json:"error"`
}

// RedfishErrorDetail contains the error code, message, and extended information
type RedfishErrorDetail struct {
	Code         string                `json:"code"`
	Message      string                `json:"message"`
	ExtendedInfo []RedfishExtendedInfo `json:"@Message.ExtendedInfo,omitempty"`
}

// RedfishExtendedInfo provides additional error details
type RedfishExtendedInfo struct {
	MessageID  string `json:"MessageId"`
	Message    string `json:"Message"`
	Severity   string `json:"Severity,omitempty"`
	Resolution string `json:"Resolution,omitempty"`
}

// RedfishError creates and sends a Redfish-compliant error response
func RedfishError(c *gin.Context, statusCode int, message string, messageID string) {
	c.JSON(statusCode, RedfishErrorResponse{
		Error: RedfishErrorDetail{
			Code:    fmt.Sprintf("Base.1.0.%s", messageID),
			Message: message,
			ExtendedInfo: []RedfishExtendedInfo{
				{
					MessageID: messageID,
					Message:   message,
					Severity:  "Critical",
				},
			},
		},
	})
}
