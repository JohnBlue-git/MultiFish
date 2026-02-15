package providers

import "fmt"

// Provider is the base interface that all providers must implement
// This enables type checking and provider identification
type Provider interface {
	// TypeName returns a human-readable name for this provider (e.g., "Redfish", "Extended", "Dell")
	TypeName() string
	
	// Supports checks if this provider can handle the given resource instance
	Supports(v interface{}) bool
	
	// SupportsCollection checks if this provider can handle the given resource collection
	SupportsCollection(v interface{}) bool
}

// GenericRegistry is a type-safe registry for any provider type
// It uses Go generics to provide compile-time type safety while maintaining flexibility
type GenericRegistry[T Provider] struct {
	providers []T
}

// NewGenericRegistry creates a new generic provider registry
// Example usage:
//   managerRegistry := NewGenericRegistry[ManagerProvider]()
//   chassisRegistry := NewGenericRegistry[ChassisProvider]()
func NewGenericRegistry[T Provider]() *GenericRegistry[T] {
	return &GenericRegistry[T]{
		providers: make([]T, 0),
	}
}

// Register adds a new provider to the registry
// Providers are checked in the order they are registered
// Register more specific providers before more general ones
func (r *GenericRegistry[T]) Register(provider T) {
	r.providers = append(r.providers, provider)
}

// FindProvider finds the appropriate provider for the given resource instance
// Returns the first provider that supports the given instance
// Returns error if no matching provider is found
func (r *GenericRegistry[T]) FindProvider(v interface{}) (T, error) {
	for _, provider := range r.providers {
		if provider.Supports(v) {
			return provider, nil
		}
	}
	var zero T
	return zero, fmt.Errorf("no provider found for type %T", v)
}

// FindCollectionProvider finds the appropriate provider for the given resource collection
// Returns the first provider that supports the given collection
// Returns error if no matching provider is found
func (r *GenericRegistry[T]) FindCollectionProvider(v interface{}) (T, error) {
	for _, provider := range r.providers {
		if provider.SupportsCollection(v) {
			return provider, nil
		}
	}
	var zero T
	return zero, fmt.Errorf("no provider found for collection type %T", v)
}

// GetProviders returns all registered providers
// Useful for debugging or listing available providers
func (r *GenericRegistry[T]) GetProviders() []T {
	return r.providers
}

// Count returns the number of registered providers
func (r *GenericRegistry[T]) Count() int {
	return len(r.providers)
}
