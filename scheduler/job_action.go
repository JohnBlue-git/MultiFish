package scheduler

import (
	"fmt"

	"multifish/utility"
	extendprovider "multifish/providers/extend"
	redfishprovider "multifish/providers/redfish"
)

// ========== Action Execution ==========

// ActionType represents supported action types
type ActionType string

const (
	// Manager-related actions
	ActionPatchProfile       ActionType = "PatchProfile"
	ActionPatchManager       ActionType = "PatchManager"
	ActionPatchFanController ActionType = "PatchFanController"
	ActionPatchFanZone       ActionType = "PatchFanZone"
	ActionPatchPidController ActionType = "PatchPidController"

	// Future actions can be added here
	// ActionReboot      ActionType = "Reboot"
	// ActionPowerOn     ActionType = "PowerOn"
	// ActionPowerOff    ActionType = "PowerOff"
)

// validateActionSupport checks if the machine supports the action
func (pv *PlatformValidator) validateActionSupport(machine interface{}, action ActionType) error {
	// For now, we assume all machines support PatchProfile and PatchManager
	// In a real implementation, you would check the machine's capabilities
	switch action {

	// Manager-related actions
	case ActionPatchManager:
		// Check if machine has managers and supports manager patching
		// This is a simplified check
		return nil
	case ActionPatchProfile:
		// Check if machine has managers and supports profile patching
		// This is a simplified check
		return nil
	case ActionPatchFanController:
		// Check if machine has managers and supports fan controller patching
		// This is a simplified check
		return nil
	case ActionPatchFanZone:
		// Check if machine has managers and supports fan zone patching
		// This is a simplified check
		return nil
	case ActionPatchPidController:
		// Check if machine has managers and supports PID controller patching
		// This is a simplified check
		return nil

	// Add other actions here with their validation logic

	default:
		return fmt.Errorf("action validation failed: action '%s' is not supported. Valid actions: %v", action, []string{"PatchProfile", "PatchManager", "PatchFanController", "PatchFanZone", "PatchPidController"})
	}
}

// MachineActionExecutor is an interface for executing actions on machines
// The main package will implement this to provide machine-specific operations
type MachineActionExecutor interface {
	// Manager-related actions
	GetManagerByService(machine interface{}, managerID string) (interface{}, error)
	PatchManager(manager interface{}, patch interface{}) error
	PatchProfile(manager interface{}, patch extendprovider.PatchProfileType) error
	PatchFanController(manager interface{}, fanControllerID string, patch *extendprovider.PatchFanControllerType) error
	PatchFanZone(manager interface{}, fanZoneID string, patch *extendprovider.PatchFanZoneType) error
	PatchPidController(manager interface{}, pidControllerID string, patch *extendprovider.PatchPidControllerType) error

	// Add other machine-specific action methods here
}

// DefaultActionExecutor provides default implementations for action execution
type DefaultActionExecutor struct {
	machineExecutor MachineActionExecutor
}

// NewDefaultActionExecutor creates a new default action executor
func NewDefaultActionExecutor(machineExecutor MachineActionExecutor) *DefaultActionExecutor {
	return &DefaultActionExecutor{
		machineExecutor: machineExecutor,
	}
}

// ActionExecutor executes specific actions on machines
type ActionExecutor interface {
	// Manager-related actions
	ExecutePatchManager(machine interface{}, managerPayloads Payload) error
	ExecutePatchProfile(machine interface{}, managerPayloads Payload) error
	ExecutePatchFanController(machine interface{}, fanControllerPayloads Payload) error
	ExecutePatchFanZone(machine interface{}, fanZonePayloads Payload) error
	ExecutePatchPidController(machine interface{}, pidControllerPayloads Payload) error

	// Add other action executors here
}

// ExecuteAction executes the specified action on a machine with the given payload
func ExecuteAction(actionExecutor ActionExecutor, action ActionType, machine interface{}, payload Payload) error {
	switch action {

	// Manager-related actions
	case ActionPatchManager:
		return actionExecutor.ExecutePatchManager(machine, payload)
	case ActionPatchProfile:
		return actionExecutor.ExecutePatchProfile(machine, payload)
	case ActionPatchFanController:
		return actionExecutor.ExecutePatchFanController(machine, payload)
	case ActionPatchFanZone:
		return actionExecutor.ExecutePatchFanZone(machine, payload)
	case ActionPatchPidController:
		return actionExecutor.ExecutePatchPidController(machine, payload)

	// Add other actions here

	default:
		return fmt.Errorf("execution failed: unsupported action '%s'. Valid actions: %v", action, []string{"PatchProfile", "PatchManager", "PatchFanController", "PatchFanZone", "PatchPidController"})
	}
}

// ========== Manager Action Execution ==========

// ExecutePatchManager executes the PatchManager action on a machine
func (dae *DefaultActionExecutor) ExecutePatchManager(machine interface{}, managerPayloads Payload) error {
	log := utility.GetLogger()

	// Type assert payload to []ExecutePatchManagerPayload
	payloads, ok := managerPayloads.([]ExecutePatchManagerPayload)
	if !ok {
		log.Error().Msgf("invalid payload type for PatchManager: expected []ExecutePatchManagerPayload, got %T", managerPayloads)
		return fmt.Errorf("execution failed: invalid payload type for PatchManager action: expected []ExecutePatchManagerPayload, got %T. Check job payload structure", managerPayloads)
	}

	// Iterate over each manager payload
	for _, mp := range payloads {
		log.Debug().
			Str("serviceIdentification", mp.Payload.ServiceIdentification).
			Str("managerID", mp.ManagerID).
			Msg("Executing PatchManager")

		// Get the manager by service using the injected executor
		manager, err := dae.machineExecutor.GetManagerByService(machine, mp.ManagerID)
		if err != nil {
			log.Error().Msgf("failed to get manager by service: %v", err)
			return fmt.Errorf("failed to retrieve manager service for manager '%s': %w. Verify machine connectivity and Redfish service availability", mp.ManagerID, err)
		}

		// Create the patch payload - convert ServiceIdentificationPayload to PatchManagerType
		patch := &redfishprovider.PatchManagerType{
			ServiceIdentification: mp.Payload.ServiceIdentification,
		}

		// Use the injected PatchManager function
		err = dae.machineExecutor.PatchManager(manager, patch)
		if err != nil {
			log.Error().Msgf("failed to patch manager %s: %v", mp.ManagerID, err)
			return fmt.Errorf("failed to patch manager '%s': %w. Check manager ID and payload values", mp.ManagerID, err)
		}
	}

	return nil
}

// ExecutePatchProfile executes the PatchProfile action on a machine
func (dae *DefaultActionExecutor) ExecutePatchProfile(machine interface{}, managerPayloads Payload) error {
	log := utility.GetLogger()

	// Type assert payload to []ExecutePatchProfilePayload
	payloads, ok := managerPayloads.([]ExecutePatchProfilePayload)
	if !ok {
		log.Error().Msgf("invalid payload type for PatchProfile: expected []ExecutePatchProfilePayload, got %T", managerPayloads)
		return fmt.Errorf("execution failed: invalid payload type for PatchProfile action: expected []ExecutePatchProfilePayload, got %T. Check job payload structure", managerPayloads)
	}

	// Iterate over each manager payload
	for _, mp := range payloads {
		log.Debug().
			Str("profile", mp.Payload.Profile).
			Str("managerID", mp.ManagerID).
			Msg("Executing PatchProfile")

		// Get the manager by service using the injected executor
		manager, err := dae.machineExecutor.GetManagerByService(machine, mp.ManagerID)
		if err != nil {
			log.Error().Msgf("failed to get manager by service: %v", err)
			return fmt.Errorf("failed to retrieve manager service for manager '%s': %w. Verify machine connectivity and Redfish service availability", mp.ManagerID, err)
		}

		// Create the patch payload
		patch := extendprovider.PatchProfileType{
			Profile: mp.Payload.Profile,
		}

		// Use the injected PatchProfile function
		err = dae.machineExecutor.PatchProfile(manager, patch)
		if err != nil {
			log.Error().Msgf("failed to patch profile for manager %s: %v", mp.ManagerID, err)
			return fmt.Errorf("failed to patch profile '%s' for manager '%s': %w. Verify profile value is valid", mp.Payload.Profile, mp.ManagerID, err)
		}
	}

	return nil
}

// ExecutePatchFanController executes the PatchFanController action on a machine
func (dae *DefaultActionExecutor) ExecutePatchFanController(machine interface{}, fanControllerPayloads Payload) error {
	log := utility.GetLogger()

	// Type assert payload to []ExecutePatchFanControllerPayload
	payloads, ok := fanControllerPayloads.([]ExecutePatchFanControllerPayload)
	if !ok {
		log.Error().Msgf("invalid payload type for PatchFanController: expected []ExecutePatchFanControllerPayload, got %T", fanControllerPayloads)
		return fmt.Errorf("execution failed: invalid payload type for PatchFanController action: expected []ExecutePatchFanControllerPayload, got %T. Check job payload structure", fanControllerPayloads)
	}
	
	// Iterate over each fan controller payload
	for _, fp := range payloads {
		log.Debug().
			Str("managerID", fp.ManagerID).
			Str("fanControllerID", fp.FanControllerID).
			Msg("Executing PatchFanController")
		
		// Get the manager by service using the injected executor
		manager, err := dae.machineExecutor.GetManagerByService(machine, fp.ManagerID)
		if err != nil {
			log.Error().Msgf("failed to get manager by service: %v", err)
			return fmt.Errorf("failed to retrieve manager service for manager '%s': %w. Verify machine connectivity and Redfish service availability", fp.ManagerID, err)
		}

		// Use the injected PatchFanController function
		err = dae.machineExecutor.PatchFanController(manager, fp.FanControllerID, &fp.Payload)
		if err != nil {
			log.Error().Msgf("failed to patch fan controller %s for manager %s: %v", fp.FanControllerID, fp.ManagerID, err)
			return fmt.Errorf("failed to patch fan controller '%s' for manager '%s': %w. Check controller ID and configuration", fp.FanControllerID, fp.ManagerID, err)
		}
	}

	return nil
}

func (dae *DefaultActionExecutor) ExecutePatchFanZone(machine interface{}, fanZonePayloads Payload) error {
	log := utility.GetLogger()

	// Type assert payload to []ExecutePatchFanZonePayload
	payloads, ok := fanZonePayloads.([]ExecutePatchFanZonePayload)
	if !ok {
		log.Error().Msgf("invalid payload type for PatchFanZone: expected []ExecutePatchFanZonePayload, got %T", fanZonePayloads)
		return fmt.Errorf("execution failed: invalid payload type for PatchFanZone action: expected []ExecutePatchFanZonePayload, got %T. Check job payload structure", fanZonePayloads)
	}

	// Iterate over each fan zone payload
	for _, fz := range payloads {
		log.Debug().
			Str("managerID", fz.ManagerID).
			Str("fanZoneID", fz.FanZoneID).
			Msg("Executing PatchFanZone")

		// Get the manager by service using the injected executor
		manager, err := dae.machineExecutor.GetManagerByService(machine, fz.ManagerID)
		if err != nil {
			log.Error().Msgf("failed to get manager by service: %v", err)
			return fmt.Errorf("failed to retrieve manager service for manager '%s': %w. Verify machine connectivity and Redfish service availability", fz.ManagerID, err)
		}

		// Use the injected PatchFanZone function
		err = dae.machineExecutor.PatchFanZone(manager, fz.FanZoneID, &fz.Payload)
		if err != nil {
			log.Error().Msgf("failed to patch fan zone %s for manager %s: %v", fz.FanZoneID, fz.ManagerID, err)
			return fmt.Errorf("failed to patch fan zone '%s' for manager '%s': %w. Verify zone ID and settings", fz.FanZoneID, fz.ManagerID, err)
		}
	}

	return nil
}

func (dae *DefaultActionExecutor) ExecutePatchPidController(machine interface{}, pidControllerPayloads Payload) error {
	log := utility.GetLogger()

	// Type assert payload to []ExecutePatchPidControllerPayload
	payloads, ok := pidControllerPayloads.([]ExecutePatchPidControllerPayload)
	if !ok {
		log.Error().Msgf("invalid payload type for PatchPidController: expected []ExecutePatchPidControllerPayload, got %T", pidControllerPayloads)
		return fmt.Errorf("execution failed: invalid payload type for PatchPidController action: expected []ExecutePatchPidControllerPayload, got %T. Check job payload structure", pidControllerPayloads)
	}

	// Iterate over each PID controller payload
	for _, pc := range payloads {
		log.Debug().
			Str("managerID", pc.ManagerID).
			Str("pidControllerID", pc.PidControllerID).
			Msg("Executing PatchPidController")

		// Get the manager by service using the injected executor
		manager, err := dae.machineExecutor.GetManagerByService(machine, pc.ManagerID)
		if err != nil {
			log.Error().Msgf("failed to get manager by service: %v", err)
			return fmt.Errorf("failed to retrieve manager service for manager '%s': %w. Verify machine connectivity and Redfish service availability", pc.ManagerID, err)
		}

		// Use the injected PatchPidController function
		err = dae.machineExecutor.PatchPidController(manager, pc.PidControllerID, &pc.Payload)
		if err != nil {
			log.Error().Msgf("failed to patch PID controller %s for manager %s: %v", pc.PidControllerID, pc.ManagerID, err)
			return fmt.Errorf("failed to patch PID controller '%s' for manager '%s': %w. Check PID parameters", pc.PidControllerID, pc.ManagerID, err)
		}
	}

	return nil
}
