package utility

import (
	"errors"
	"net/http"
	"testing"
)

// TestResponseError tests the ResponseError structure
func TestResponseError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		err        error
		message    string
	}{
		{
			name:       "not found error",
			statusCode: http.StatusNotFound,
			err:        errors.New("resource not found"),
			message:    "ResourceNotFound",
		},
		{
			name:       "bad request error",
			statusCode: http.StatusBadRequest,
			err:        errors.New("invalid input"),
			message:    "InvalidProperty",
		},
		{
			name:       "internal error",
			statusCode: http.StatusInternalServerError,
			err:        errors.New("internal server error"),
			message:    "InternalError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			respErr := &ResponseError{
				StatusCode: tt.statusCode,
				Error:      tt.err,
				Message:    tt.message,
			}

			if respErr.StatusCode != tt.statusCode {
				t.Errorf("ResponseError.StatusCode = %v, want %v", respErr.StatusCode, tt.statusCode)
			}

			if respErr.Error != tt.err {
				t.Errorf("ResponseError.Error = %v, want %v", respErr.Error, tt.err)
			}

			if respErr.Message != tt.message {
				t.Errorf("ResponseError.Message = %v, want %v", respErr.Message, tt.message)
			}
		})
	}
}

// TestResponseError_NilError tests ResponseError with nil error
func TestResponseError_NilError(t *testing.T) {
	respErr := &ResponseError{
		StatusCode: http.StatusOK,
		Error:      nil,
		Message:    "Success",
	}

	if respErr.Error != nil {
		t.Errorf("ResponseError.Error = %v, want nil", respErr.Error)
	}
}
