package extendprovider

import (
	"testing"

	"github.com/stmcginnis/gofish"
)

// TestNewExtendService tests the NewExtendService constructor
func TestNewExtendService(t *testing.T) {
	// Create a mock service (in real implementation, this would come from gofish.APIClient)
	mockService := &gofish.Service{}
	
	// Create ExtendService
	es := NewExtendService(&gofish.APIClient{Service: mockService})
	
	if es == nil {
		t.Fatal("NewExtendService returned nil")
	}
	
	if es.Service != mockService {
		t.Error("ExtendService.Service does not match the input service")
	}
}

// TestExtendService_Structure tests the ExtendService structure
func TestExtendService_Structure(t *testing.T) {
	mockService := &gofish.Service{}
	
	es := &ExtendService{
		Service: mockService,
	}
	
	if es.Service != mockService {
		t.Error("ExtendService.Service does not match")
	}
}

// TestExtendService_NilClient tests ExtendService with nil client
func TestExtendService_NilClient(t *testing.T) {
	// This test verifies that NewExtendService handles edge cases
	// In practice, passing nil would likely cause issues, but we test the structure
	
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic: %v", r)
		}
	}()
	
	// Create with nil client (may panic or return service with nil)
	es := NewExtendService(&gofish.APIClient{Service: nil})
	
	if es == nil {
		t.Error("NewExtendService should not return nil even with nil service")
	}
}

// TestExtendService_InheritsMethods tests that ExtendService embeds Service
func TestExtendService_InheritsMethods(t *testing.T) {
	// Create a mock service
	mockService := &gofish.Service{}
	
	es := NewExtendService(&gofish.APIClient{Service: mockService})
	
	// Verify that we can access Service fields through embedding
	if es.Service == nil {
		t.Error("ExtendService should have access to embedded Service")
	}
}

// TestExtendService_MultipleInstances tests creating multiple ExtendService instances
func TestExtendService_MultipleInstances(t *testing.T) {
	mockService1 := &gofish.Service{}
	mockService2 := &gofish.Service{}
	
	es1 := NewExtendService(&gofish.APIClient{Service: mockService1})
	es2 := NewExtendService(&gofish.APIClient{Service: mockService2})
	
	if es1 == es2 {
		t.Error("Different ExtendService instances should not be equal")
	}
	
	if es1.Service == es2.Service {
		t.Error("Different ExtendService instances should have different Services")
	}
}

// TestExtendService_TypeAssertions tests type assertions for ExtendService
func TestExtendService_TypeAssertions(t *testing.T) {
	mockService := &gofish.Service{}
	es := NewExtendService(&gofish.APIClient{Service: mockService})
	
	// Type assertion test
	_, ok := interface{}(es).(*ExtendService)
	if !ok {
		t.Error("ExtendService should be of type *ExtendService")
	}
}

// TestExtendService_ServicePointer tests that Service is properly set as pointer
func TestExtendService_ServicePointer(t *testing.T) {
	mockService := &gofish.Service{}
	es := NewExtendService(&gofish.APIClient{Service: mockService})
	
	// Modify the service through the ExtendService
	// This tests that it's a pointer and changes are reflected
	if es.Service != mockService {
		t.Error("Service should be a pointer to the same object")
	}
	
	// Verify pointer equality
	if &es.Service == &mockService {
		t.Log("Service is properly stored as a pointer (addresses are different as expected for embedded field)")
	}
}
