# Providers Module

## Overview

The `providers/` package implements a flexible, extensible **Generic Provider Pattern** for managing different types of BMC/Redfish implementations. It uses Go generics to provide type-safe, reusable provider registries that support both standard Redfish (Base) and extended OEM implementations (Extend).

> **📚 For comprehensive documentation on the Generic Provider Pattern, see [PROVIDER_PATTERN.md](./PROVIDER_PATTERN.md)**

## Quick Links

- **[Provider Pattern Guide](./PROVIDER_PATTERN.md)** - Complete guide to the generic provider pattern
- **[Adding New Registry Types](./PROVIDER_PATTERN.md#adding-a-new-registry-type)** - Step-by-step guide
- **[Vendor-Specific Providers](./PROVIDER_PATTERN.md#example-adding-vendor-specific-provider)** - Examples for Dell, HP, etc.

## Why Provider Pattern?

The provider pattern solves several key challenges:
- **Vendor Abstraction**: Support multiple BMC vendors (OpenBMC, Dell iDRAC, HPE iLO, etc.)
- **Feature Detection**: Automatically detect and use vendor-specific features
- **Extensibility**: Easy to add new providers without modifying existing code
- **Type Safety**: Go generics provide compile-time type guarantees
- **Fallback Mechanism**: Gracefully degrade to standard features when OEM features unavailable
- **Reusability**: Generic registry works for any resource type (Managers, Chassis, Systems, etc.)

## Architecture

```
providers/
├── generic_registry.go        # Generic registry implementation (NEW!)
├── manager_provider.go        # Manager provider interface + ManagerRegistry
├── provider_registry.go       # Backward compatibility (deprecated)
├── README.md                  # This file (quick reference)
├── PROVIDER_PATTERN.md        # Complete provider pattern guide
├── redfish/                   # Standard Redfish provider
│   ├── BaseManager.go         # Base Redfish manager implementation
│   └── BaseManager_test.go    # Base manager tests
└── extend/                    # OEM Extended provider
    ├── ExtendManager.go       # Extended manager with OEM features
    ├── ExtendManager_test.go  # Extended manager tests
    ├── ExtendService.go       # Extended service implementation
    └── ExtendService_test.go  # Extended service tests
```

## Core Components

### 1. Generic Registry (`generic_registry.go`) - NEW! ✨

Type-safe, reusable registry using Go generics:

```go
// Base interface all providers implement
type Provider interface {
    TypeName() string
    Supports(v interface{}) bool
    SupportsCollection(v interface{}) bool
}

// Generic registry works with any provider type
type GenericRegistry[T Provider] struct {
    providers []T
}

// Example usage:
managerRegistry := NewGenericRegistry[ManagerProvider]()
chassisRegistry := NewGenericRegistry[ChassisProvider]()
systemRegistry := NewGenericRegistry[SystemProvider]()
```

**Key Benefits:**
- ✅ Type-safe at compile time
- ✅ Reusable for any resource type
- ✅ No code duplication
- ✅ Easy to extend

### 2. Provider Registry (`provider_registry.go`)

Backward compatibility file - provides deprecated aliases:

```go
// Deprecated: Use ManagerRegistry instead
type ProviderRegistry = ManagerRegistry

// Deprecated: Use NewManagerRegistry instead
func NewProviderRegistry() *ManagerRegistry {
    return NewManagerRegistry()
}
```

**Note:** This file exists only for backward compatibility. All new code should use `ManagerRegistry` from `manager_provider.go`.

### 3. Manager Provider Interface (`manager_provider.go`)

Defines the contract all manager providers must implement, plus the ManagerRegistry:

```go
// Manager provider interface
type ManagerProvider interface {
    Provider  // Embeds base Provider interface
    
    // Manager operations
    GetManagerResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError)
    PatchManager(manager interface{}, updates interface{}) *utility.ResponseError
    
    // Profile operations (if supported)
    GetProfileResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError)
    PatchProfile(manager interface{}, updates interface{}) *utility.ResponseError
    
    // OEM operations (if supported)
    GetFanControllerCollectionResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError)
    // ... more operations
}

// Manager-specific registry
type ManagerRegistry struct {
    *GenericRegistry[ManagerProvider]
}

func NewManagerRegistry() *ManagerRegistry {
    return &ManagerRegistry{
        GenericRegistry: NewGenericRegistry[ManagerProvider](),
    }
}
```

**How it works:**
1. Providers are registered in priority order
2. Registry checks each provider's `Supports()` method
3. Returns first matching provider
4. Falls back to next provider if not supported

**Example:**
```go
// In init()
managerProviders = providers.NewManagerRegistry()

// Register in priority order (most specific first)
managerProviders.Register(&extendprovider.ExtendProvider{})
managerProviders.Register(&redfishprovider.RedfishProvider{})
```

Defines the contract all providers must implement.

```go
type ManagerProvider interface {
    // Supports checks if this provider can handle the given manager
    Supports(manager interface{}) bool
    
    // Manager operations
    GetManager(manager interface{}) (interface{}, error)
    PatchManager(manager interface{}, payload interface{}) error
    
    // Profile operations (if supported)
    GetProfile(manager interface{}) (interface{}, error)
    PatchProfile(manager interface{}, payload interface{}) error
    
    // OEM operations (if supported)
    GetOemData(manager interface{}) (interface{}, error)
    GetFanControllers(manager interface{}) (interface{}, error)
    // ... more OEM operations
}
```

**Why separate methods?**
- Clear capability boundaries
- Optional feature support
- Graceful degradation

## Provider Implementations

### Redfish Provider (`redfish/`)

Standard Redfish implementation for base BMC functionality.

#### BaseManager (`BaseManager.go`)

Implements core Redfish manager operations.

**Features:**
- Manager metadata retrieval
- Basic manager patching
- Standard Redfish compliance

**Supported Operations:**
```go
// Get manager details
manager, err := provider.GetManager(managerInterface)

// Patch manager properties
patch := &PatchManagerType{
    ServiceIdentification: "New ID",
}
err := provider.PatchManager(manager, patch)
```

**Allowed PATCH Fields:**
```go
func ManagerAllowedPatchFields() []string {
    return []string{
        "ServiceIdentification",
        // Add more standard Redfish fields
    }
}
```

**When to use:**
- Standard Redfish-compliant BMCs
- Minimal configuration changes
- Cross-vendor compatibility required

#### PatchManagerType

```go
type PatchManagerType struct {
    ServiceIdentification string `json:"ServiceIdentification,omitempty"`
}
```

### Extend Provider (`extend/`)

OpenBMC-specific implementation with OEM extensions.

#### ExtendManager (`ExtendManager.go`)

Extended manager with OpenBMC-specific features.

**Additional Features:**
- Profile management (Performance, Balanced, PowerSaver, Custom)
- Fan controller configuration
- Fan zone management
- PID controller tuning

**Structure:**
```go
type ExtendManager struct {
    *redfish.Manager  // Embeds standard manager
    OEM               *OEM
}

type OEM struct {
    OpenBmc *OpenBmc `json:"OpenBmc"`
}

type OpenBmc struct {
    Profile        *Profile
    FanControllers *FanControllers
    FanZones       *FanZones
    PidControllers *PidControllers
}
```

#### Profile Management

**Why:** Provides pre-configured thermal profiles for different workload types.

**Profiles:**
- **Performance**: Maximum performance, higher power/heat
- **Balanced**: Optimal balance of performance and efficiency
- **PowerSaver**: Minimize power consumption
- **Custom**: User-defined settings

**Operations:**
```go
// Get current profile
profile, err := provider.GetProfile(manager)

// Change profile
patch := extendprovider.PatchProfileType{
    Profile: "Performance",
}
err := provider.PatchProfile(manager, &patch)
```

**Validation:**
```go
var ProfileAllowlist = []string{
    "Performance",
    "Balanced",
    "PowerSaver",
    "Custom",
}
```

#### Fan Controller Management

**Why:** Fine-tune fan behavior for specific thermal requirements.

**Configuration Parameters:**
```go
type PatchFanControllerType struct {
    Multiplier *float64 `json:"Multiplier,omitempty"`  // Fan speed multiplier
    StepDown   *int     `json:"StepDown,omitempty"`    // Speed decrease step
    StepUp     *int     `json:"StepUp,omitempty"`      // Speed increase step
}
```

**Use Cases:**
- Adjust fan response to temperature changes
- Balance noise vs. cooling
- Prevent rapid fan speed fluctuations

**Example:**
```go
patch := &extendprovider.PatchFanControllerType{
    Multiplier: &1.2,      // 20% faster
    StepDown:   &2,        // Slower decrease
    StepUp:     &5,        // Faster increase
}
err := provider.PatchFanController(manager, "cpu_fan_controller", patch)
```

#### Fan Zone Management

**Why:** Define thermal zones with specific cooling requirements.

**Configuration:**
```go
type PatchFanZoneType struct {
    FailSafePercent  *float64 `json:"FailSafePercent,omitempty"`   // Emergency speed
    MinThermalOutput *float64 `json:"MinThermalOutput,omitempty"`  // Minimum fan speed
}
```

**Example:**
```go
patch := &extendprovider.PatchFanZoneType{
    FailSafePercent:  &100.0,  // Full speed on failure
    MinThermalOutput: &30.0,   // Never below 30%
}
err := provider.PatchFanZone(manager, "cpu_zone", patch)
```

#### PID Controller Management

**Why:** Advanced thermal control using PID (Proportional-Integral-Derivative) algorithms.

**Configuration:**
```go
type PatchPidControllerType struct {
    FFGainCoefficient *float64 `json:"FFGainCoefficient,omitempty"`
    FFOffCoefficient  *float64 `json:"FFOffCoefficient,omitempty"`
    ICoefficient      *float64 `json:"ICoefficient,omitempty"`
    ILimitMax         *float64 `json:"ILimitMax,omitempty"`
    ILimitMin         *float64 `json:"ILimitMin,omitempty"`
    OutLimitMax       *float64 `json:"OutLimitMax,omitempty"`
    OutLimitMin       *float64 `json:"OutLimitMin,omitempty"`
    PCoefficient      *float64 `json:"PCoefficient,omitempty"`
    SetPoint          *float64 `json:"SetPoint,omitempty"`
    SlewNeg           *float64 `json:"SlewNeg,omitempty"`
    SlewPos           *float64 `json:"SlewPos,omitempty"`
}
```

**PID Parameters Explained:**
- **PCoefficient**: Proportional gain (response to current error)
- **ICoefficient**: Integral gain (response to accumulated error)
- **FFGainCoefficient**: Feed-forward gain
- **SetPoint**: Target temperature
- **SlewPos/SlewNeg**: Maximum rate of change

**Example:**
```go
patch := &extendprovider.PatchPidControllerType{
    PCoefficient: &0.8,
    ICoefficient: &0.1,
    SetPoint:     &75.0,  // Target 75°C
}
err := provider.PatchPidController(manager, "cpu_temp_controller", patch)
```

#### ExtendService (`ExtendService.go`)

Service wrapper that provides extended manager instances.

**Why:** Ensures manager objects include OEM extensions.

```go
type ExtendService struct {
    service *gofish.Service
}

func (es *ExtendService) Managers() ([]*ExtendManager, error)
```

## Provider Selection Flow

```
Request comes in
       ↓
Get manager instance
       ↓
Call Registry.FindProvider(manager)
       ↓
For each registered provider:
   - Check provider.Supports(manager)
   - If true, use this provider
   - If false, try next provider
       ↓
Provider executes operation
```

**Example:**
```go
// In handler/handleManager.go
provider := ManagerProviders.FindProvider(manager)
if provider == nil {
    return error("no suitable provider found")
}

// Provider automatically handles correct type
result, err := provider.GetProfile(manager)
```

## Adding a New Provider

### Step 1: Implement ManagerProvider Interface

```go
package dellprovider

type DellProvider struct{}

func (p *DellProvider) Supports(manager interface{}) bool {
    // Check if this is a Dell iDRAC
    if m, ok := manager.(*redfish.Manager); ok {
        return m.Manufacturer == "Dell" 
    }
    return false
}

func (p *DellProvider) GetManager(manager interface{}) (interface{}, error) {
    // Implementation
}

// Implement other required methods...
```

### Step 2: Register Provider

```go
// In main.go or handler init()
managerProviders.Register(&dellprovider.DellProvider{})
managerProviders.Register(&extendprovider.ExtendProvider{})
managerProviders.Register(&redfishprovider.RedfishProvider{})
```

### Step 3: Define Provider-Specific Types

```go
type DellManager struct {
    *redfish.Manager
    OEM *DellOEM
}

type DellOEM struct {
    Dell *DellExtensions
}
```

## Testing Providers

### Unit Tests

```go
func TestProviderSupports(t *testing.T) {
    provider := &ExtendProvider{}
    
    tests := []struct {
        name    string
        manager interface{}
        want    bool
    }{
        {
            name:    "extended manager",
            manager: &ExtendManager{},
            want:    true,
        },
        {
            name:    "base manager",
            manager: &redfish.Manager{},
            want:    false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := provider.Supports(tt.manager)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Integration Tests

```go
func TestProviderRegistry(t *testing.T) {
    registry := NewProviderRegistry()
    registry.Register(&ExtendProvider{})
    registry.Register(&RedfishProvider{})
    
    manager := &ExtendManager{}
    provider := registry.FindProvider(manager)
    
    assert.NotNil(t, provider)
    assert.IsType(t, &ExtendProvider{}, provider)
}
```

## Best Practices

### 1. Provider Registration Order

Register most specific providers first:
```go
// Correct order
managerProviders.Register(&DellProvider{})     // Most specific
managerProviders.Register(&ExtendProvider{})   // OEM extensions
managerProviders.Register(&RedfishProvider{})  // Fallback
```

### 2. Graceful Degradation

```go
provider := managerProviders.FindProvider(manager)
if provider == nil {
    // Fallback to basic operations
    return handleBasicManager(manager)
}
```

### 3. Type Assertions

```go
func (p *ExtendProvider) GetProfile(manager interface{}) (interface{}, error) {
    extManager, ok := manager.(*ExtendManager)
    if !ok {
        return nil, fmt.Errorf("unsupported manager type: %T", manager)
    }
    // Use extManager safely
}
```

### 4. Validation

```go
func (p *ExtendProvider) PatchProfile(manager interface{}, payload interface{}) error {
    patch, ok := payload.(*PatchProfileType)
    if !ok {
        return fmt.Errorf("invalid payload type")
    }
    
    if !IsValidProfile(patch.Profile) {
        return fmt.Errorf("invalid profile: %s", patch.Profile)
    }
    
    // Apply patch
}
```

## Common Patterns

### Provider Selection

```go
func getProviderForManager(manager interface{}) ManagerProvider {
    provider := managerProviders.FindProvider(manager)
    if provider == nil {
        log.Printf("No provider found for manager type: %T", manager)
        return nil
    }
    return provider
}
```

### Feature Detection

```go
func supportsProfiles(manager interface{}) bool {
    provider := managerProviders.FindProvider(manager)
    if provider == nil {
        return false
    }
    
    _, err := provider.GetProfile(manager)
    return err == nil
}
```

### Safe Type Conversion

```go
func toExtendManager(manager interface{}) (*ExtendManager, error) {
    if extMgr, ok := manager.(*ExtendManager); ok {
        return extMgr, nil
    }
    return nil, fmt.Errorf("manager is not ExtendManager type")
}
```

## Quick Start: Adding New Registry Types

Want to add support for Chassis, Systems, or other Redfish resources? Here's a minimal example:

### 1. Define the Provider Interface

```go
// File: providers/chassis_provider.go
package providers

type ChassisProvider interface {
    Provider  // Embed base Provider interface
    
    GetChassisResponse(chassis interface{}, machineID, chassisID string) (gin.H, error)
    PatchChassis(chassis interface{}, updates interface{}) error
    // ... add more methods
}
```

### 2. Create Registry Type Alias

```go
// File: providers/chassis_registry.go
package providers

// ChassisRegistry uses the generic registry
type ChassisRegistry = GenericRegistry[ChassisProvider]

func NewChassisRegistry() *ChassisRegistry {
    return NewGenericRegistry[ChassisProvider]()
}

// Add delegation methods
func (r *ChassisRegistry) GetChassisResponse(chassis interface{}, machineID, chassisID string) (gin.H, error) {
    provider, err := r.FindProvider(chassis)
    if err != nil {
        return nil, err
    }
    return provider.GetChassisResponse(chassis, machineID, chassisID)
}
```

### 3. Implement Concrete Provider

```go
// File: providers/redfish/BaseChassis.go
package redfish

type RedfishChassisProvider struct{}

func (p *RedfishChassisProvider) TypeName() string { return "RedfishChassis" }
func (p *RedfishChassisProvider) Supports(v interface{}) bool {
    _, ok := v.(*redfish.Chassis)
    return ok
}
// ... implement ChassisProvider methods
```

### 4. Register and Use

```go
// In your application
chassisProviders := NewChassisRegistry()
chassisProviders.Register(&RedfishChassisProvider{})

// Use in handlers
response, err := chassisProviders.GetChassisResponse(chassis, machineID, chassisID)
```
