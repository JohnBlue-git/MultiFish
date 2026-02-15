package scheduler

import (
	"encoding/json"
	"fmt"

	"multifish/utility"
	redfish "multifish/providers/redfish"
	extendprovider "multifish/providers/extend"
)

// structToMap converts a struct to map[string]interface{} for validation
func structToMap(v interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal struct: %v", err)
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %v", err)
	}
	
	return result, nil
}

// validatePayloadSupport checks if the machine supports the payload
func (pv *PlatformValidator) validatePayloadSupport(machine interface{}, action ActionType, payload Payload) error {
	switch action {

	// Manager-related actions
	case ActionPatchManager:
		// Type assert payload to []ExecutePatchManagerPayload
		payloads, ok := payload.([]ExecutePatchManagerPayload)
		if !ok {
			return fmt.Errorf("payload validation failed for action '%s': invalid payload type for PatchManager, expected []ExecutePatchManagerPayload, got %T. Check JSON payload structure", action, payload)
		}
		if len(payloads) == 0 {
			return fmt.Errorf("payload validation failed for action '%s': at least one manager patch payload is required in the array. Add manager configurations to the payload", action)
		}
		
		// Validate each manager patch payload
		for i, mp := range payloads {
			if mp.ManagerID == "" {
				return fmt.Errorf("payload validation failed at index %d: ManagerID is required and cannot be empty. Specify a valid manager identifier", i)
			}
			if mp.Payload.ServiceIdentification == "" {
				return fmt.Errorf("payload validation failed for ManagerID '%s' at index %d: ServiceIdentification cannot be empty. Specify a service identification value", mp.ManagerID, i)
			}
		}
		return nil
	case ActionPatchProfile:
		// Type assert payload to []ExecutePatchProfilePayload
		payloads, ok := payload.([]ExecutePatchProfilePayload)
		if !ok {
			return fmt.Errorf("payload validation failed for action '%s': invalid payload type for PatchProfile, expected []ExecutePatchProfilePayload, got %T. Check JSON payload structure", action, payload)
		}
		if len(payloads) == 0 {
			return fmt.Errorf("payload validation failed for action '%s': at least one manager payload is required in the array. Add manager configurations to the payload", action)
		}
		
		// Validate each manager payload
		for i, mp := range payloads {
			if mp.ManagerID == "" {
				return fmt.Errorf("payload validation failed at index %d: ManagerID is required and cannot be empty. Specify a valid manager identifier", i)
			}
			if mp.Payload.Profile == "" {
				return fmt.Errorf("payload validation failed for ManagerID '%s' at index %d: profile cannot be empty. Specify a profile value (Performance, Balanced, PowerSaver, Custom)", mp.ManagerID, i)
			}
		}
		return nil
	case ActionPatchFanController:
		// Type assert payload to []ExecutePatchFanControllerPayload
		payloads, ok := payload.([]ExecutePatchFanControllerPayload)
		if !ok {
			return fmt.Errorf("payload validation failed for action '%s': invalid payload type for PatchFanController, expected []ExecutePatchFanControllerPayload, got %T. Check JSON payload structure", action, payload)
		}
		if len(payloads) == 0 {
			return fmt.Errorf("payload validation failed for action '%s': at least one fan controller payload is required in the array. Add fan controller configurations to the payload", action)
		}
		
		// Validate each fan controller payload
		for i, fp := range payloads {
			if fp.ManagerID == "" {
				return fmt.Errorf("payload validation failed at index %d: ManagerID is required and cannot be empty. Specify a valid manager identifier", i)
			}
			if fp.FanControllerID == "" {
				return fmt.Errorf("payload validation failed for ManagerID '%s' at index %d: FanControllerID cannot be empty. Specify a valid fan controller identifier", fp.ManagerID, i)
			}
		}
		return nil
	case ActionPatchFanZone:
		// Type assert payload to []ExecutePatchFanZonePayload
		payloads, ok := payload.([]ExecutePatchFanZonePayload)
		if !ok {
			return fmt.Errorf("payload validation failed for action '%s': invalid payload type for PatchFanZone, expected []ExecutePatchFanZonePayload, got %T. Check JSON payload structure", action, payload)
		}
		if len(payloads) == 0 {
			return fmt.Errorf("payload validation failed for action '%s': at least one fan zone payload is required in the array. Add fan zone configurations to the payload", action)
		}
		
		// Validate each fan zone payload
		for i, fz := range payloads {
			if fz.ManagerID == "" {
				return fmt.Errorf("payload validation failed at index %d: ManagerID is required and cannot be empty. Specify a valid manager identifier", i)
			}
			if fz.FanZoneID == "" {
				return fmt.Errorf("payload validation failed for ManagerID '%s' at index %d: FanZoneID cannot be empty. Specify a valid fan zone identifier", fz.ManagerID, i)
			}
		}
		return nil
	case ActionPatchPidController:
		// Type assert payload to []ExecutePatchPidControllerPayload
		payloads, ok := payload.([]ExecutePatchPidControllerPayload)
		if !ok {
			return fmt.Errorf("payload validation failed for action '%s': invalid payload type for PatchPidController, expected []ExecutePatchPidControllerPayload, got %T. Check JSON payload structure", action, payload)
		}
		if len(payloads) == 0 {
			return fmt.Errorf("payload validation failed for action '%s': at least one PID controller payload is required in the array. Add PID controller configurations to the payload", action)
		}
		
		// Validate each PID controller payload
		for i, pc := range payloads {
			if pc.ManagerID == "" {
				return fmt.Errorf("payload validation failed at index %d: ManagerID is required and cannot be empty. Specify a valid manager identifier", i)
			}
			if pc.PidControllerID == "" {
				return fmt.Errorf("payload validation failed for ManagerID '%s' at index %d: PidControllerID cannot be empty. Specify a valid PID controller identifier", pc.ManagerID, i)
			}
		}
		return nil

	// Add other actions here with their validation logic

	default:
		return fmt.Errorf("payload validation failed: action '%s' is not supported. Supported actions: %v", action, []string{"PatchProfile", "PatchManager", "PatchFanController", "PatchFanZone", "PatchPidController"})
	}
}

// validatePayload validates the payload based on the action
func (j *JobCreateRequest) validatePayload() error {
	switch j.Action {

	// Manager-related actions
	case ActionPatchProfile:
		payload, ok := j.Payload.([]ExecutePatchProfilePayload)
		if !ok {
			return fmt.Errorf("job validation failed: invalid payload format for PatchProfile action. Expected array of ExecutePatchProfilePayload objects. See API documentation for correct structure")
		}
		return ValidateProfilePayloads(payload)
	case ActionPatchManager:
		payload, ok := j.Payload.([]ExecutePatchManagerPayload)
		if !ok {
			return fmt.Errorf("job validation failed: invalid payload format for PatchManager action. Expected array of ExecutePatchManagerPayload objects. See API documentation for correct structure")
		}
		return ValidateManagerPatchPayloads(payload)
	case ActionPatchFanController:
		payload, ok := j.Payload.([]ExecutePatchFanControllerPayload)
		if !ok {
			return fmt.Errorf("job validation failed: invalid payload format for PatchFanController action. Expected array of ExecutePatchFanControllerPayload objects. See API documentation for correct structure")
		}
		return ValidateFanControllerPayloads(payload)
	case ActionPatchFanZone:
		payload, ok := j.Payload.([]ExecutePatchFanZonePayload)
		if !ok {
			return fmt.Errorf("job validation failed: invalid payload format for PatchFanZone action. Expected array of ExecutePatchFanZonePayload objects. See API documentation for correct structure")
		}
		return ValidateFanZonePayloads(payload)
	case ActionPatchPidController:
		payload, ok := j.Payload.([]ExecutePatchPidControllerPayload)
		if !ok {
			return fmt.Errorf("job validation failed: invalid payload format for PatchPidController action. Expected array of ExecutePatchPidControllerPayload objects. See API documentation for correct structure")
		}
		return ValidatePidControllerPayloads(payload)

	// Add other actions here

	default:
		return fmt.Errorf("job validation failed: unsupported action type '%s'. Supported actions: %v. Update the 'action' field in job definition", j.Action, []string{"PatchProfile", "PatchManager", "PatchFanController", "PatchFanZone", "PatchPidController"})
	}
}

// ========== Manager Payload Validation ==========

// ExecutePatchManagerPayload represents a manager patch payload for PatchManager action
type ExecutePatchManagerPayload struct {
	ManagerID string                    `json:"ManagerID"`
	Payload   redfish.PatchManagerType 	`json:"Payload"`
}

// ValidateManagerPatchPayloads validates an array of manager patch payloads for PatchManager action
func ValidateManagerPatchPayloads(payloads Payload) error {
	// Type assert payload to []ExecutePatchManagerPayload
	managerPayloads, ok := payloads.([]ExecutePatchManagerPayload)
	if !ok {
		return fmt.Errorf("invalid payload type: expected []ExecutePatchManagerPayload, got %T", payloads)
	}
	
	if len(managerPayloads) == 0 {
		return fmt.Errorf("at least one manager patch payload is required for PatchManager action")
	}

	// Validate each manager patch payload
	seenManagers := make(map[string]bool)
	for i, mp := range managerPayloads {
		// Check for empty ManagerID
		if mp.ManagerID == "" {
			return fmt.Errorf("payload validation failed at payload[%d]: ManagerID is required and cannot be empty. Add a valid manager identifier", i)
		}

		// Check for duplicate ManagerIDs
		if seenManagers[mp.ManagerID] {
			return fmt.Errorf("payload validation failed: duplicate ManagerID '%s' found at payload[%d]. Each manager can only appear once. Remove duplicate entries", mp.ManagerID, i)
		}
		seenManagers[mp.ManagerID] = true

		// Validate payload fields - convert struct to map first
		payloadMap, err := structToMap(mp.Payload)
		if err != nil {
			return fmt.Errorf("payload validation failed at payload[%d] (ManagerID=%s): failed to convert payload to expected format: %w. Check field types and values match API schema", i, mp.ManagerID, err)
		}
		
		if errResp := utility.CheckPatchPayloadHelper(payloadMap, redfish.ManagerAllowedPatchFields()); errResp != nil {
			return fmt.Errorf("payload validation failed at payload[%d] (ManagerID=%s): %s", i, mp.ManagerID, errResp.Error.Message)
		}
	}
	return nil
}

// ExecutePatchProfilePayload represents a manager-specific payload
type ExecutePatchProfilePayload struct {
	ManagerID string         				  `json:"ManagerID"`
	Payload   extendprovider.PatchProfileType `json:"Payload"`
}

// IsValidProfile checks if a profile value is valid
func IsValidProfile(profile string) bool {
	for _, vp := range extendprovider.ProfileAllowlist {
		if profile == vp {
			return true
		}
	}
	return false
}

// ValidateProfilePayloads validates an array of manager payloads for PatchProfile action
func ValidateProfilePayloads(payloads Payload) error {
	// Type assert payload to []ExecutePatchProfilePayload
	managerPayloads, ok := payloads.([]ExecutePatchProfilePayload)
	if !ok {
		return fmt.Errorf("invalid payload type: expected []ExecutePatchProfilePayload, got %T", payloads)
	}
	
	if len(managerPayloads) == 0 {
		return fmt.Errorf("at least one manager payload is required for PatchProfile action")
	}

	// Validate each manager payload
	seenManagers := make(map[string]bool)
	for i, mp := range managerPayloads {
		// Check for empty ManagerID
		if mp.ManagerID == "" {
			return fmt.Errorf("payload validation failed at payload[%d]: ManagerID is required and cannot be empty. Add a valid manager identifier", i)
		}

		// Check for duplicate ManagerIDs
		if seenManagers[mp.ManagerID] {
			return fmt.Errorf("payload validation failed: duplicate ManagerID '%s' found at payload[%d]. Each manager can only appear once. Remove duplicate entries", mp.ManagerID, i)
		}
		seenManagers[mp.ManagerID] = true

		// Validate payload fields - convert struct to map first
		payloadMap, err := structToMap(mp.Payload)
		if err != nil {
			return fmt.Errorf("payload validation failed at payload[%d] (ManagerID=%s): failed to convert payload to expected format: %w. Check field types and values match API schema", i, mp.ManagerID, err)
		}
		
		if errResp := utility.CheckPatchPayloadHelper(payloadMap, extendprovider.ProfileAllowedPatchFields()); errResp != nil {
			return fmt.Errorf("payload validation failed at payload[%d] (ManagerID=%s): %s", i, mp.ManagerID, errResp.Error.Message)
		}
		
		// Validate profile values
		if !IsValidProfile(mp.Payload.Profile) {
			return fmt.Errorf("payload validation failed at payload[%d] (ManagerID=%s): invalid Profile value '%s'. Must be one of: Performance, Balanced, PowerSaver, Custom. Update the profile field in your payload",
				i, mp.ManagerID, mp.Payload.Profile)
		}
	}
	return nil
}

// ExecutePatchFanControllerPayload represents a fan controller patch payload
type ExecutePatchFanControllerPayload struct {
	ManagerID       string                          	  `json:"ManagerID"`
	FanControllerID string                          	  `json:"FanControllerID"`
	Payload         extendprovider.PatchFanControllerType `json:"Payload"`
}

// ValidateFanControllerPayloads validates an array of fan controller patch payloads
func ValidateFanControllerPayloads(payloads Payload) error {
	// Type assert payload to []ExecutePatchFanControllerPayload
	fanControllerPayloads, ok := payloads.([]ExecutePatchFanControllerPayload)
	if !ok {
		return fmt.Errorf("invalid payload type: expected []ExecutePatchFanControllerPayload, got %T", payloads)
	}
	
	if len(fanControllerPayloads) == 0 {
		return fmt.Errorf("at least one fan controller patch payload is required")
	}

	// Validate each fan controller patch payload
	seenControllers := make(map[string]bool)
	for i, fp := range fanControllerPayloads {
		// Check for empty ManagerID
		if fp.ManagerID == "" {
			return fmt.Errorf("payload validation failed at payload[%d]: ManagerID is required and cannot be empty. Add a valid manager identifier", i)
		}

		// Check for empty FanControllerID
		if fp.FanControllerID == "" {
			return fmt.Errorf("payload validation failed at payload[%d]: FanControllerID is required and cannot be empty. Add a valid fan controller identifier", i)
		}

		// Check for duplicate FanControllerIDs per ManagerID
		key := fmt.Sprintf("%s:%s", fp.ManagerID, fp.FanControllerID)
		if seenControllers[key] {
			return fmt.Errorf("payload validation failed: duplicate FanControllerID '%s' found for ManagerID '%s' at payload[%d]. Each fan controller can only appear once per manager. Remove duplicate entries", fp.FanControllerID, fp.ManagerID, i)
		}
		seenControllers[key] = true

		// Validate payload fields - convert struct to map first
		payloadMap, err := structToMap(fp.Payload)
		if err != nil {
			return fmt.Errorf("payload validation failed at payload[%d] (ManagerID=%s, FanControllerID=%s): failed to convert payload to expected format: %w. Check field types and values match API schema", i, fp.ManagerID, fp.FanControllerID, err)
		}

		// Validate payload fields
		if errResp := utility.CheckPatchPayloadHelper(payloadMap, extendprovider.FanControllerAllowedPatchFields()); errResp != nil {
			return fmt.Errorf("payload validation failed at payload[%d] (ManagerID=%s, FanControllerID=%s): %s", i, fp.ManagerID, fp.FanControllerID, errResp.Error.Message)
		}
	}
	return nil
}

// ExecutePatchFanZonePayload represents a fan zone patch payload
type ExecutePatchFanZonePayload struct {
	ManagerID      string                          	  `json:"ManagerID"`
	FanZoneID      string                          	  `json:"FanZoneID"`
	Payload        extendprovider.PatchFanZoneType    `json:"Payload"`
}

// ValidateFanZonePayloads validates an array of fan zone patch payloads
func ValidateFanZonePayloads(payloads Payload) error {
	// Type assert payload to []ExecutePatchFanZonePayload
	fanZonePayloads, ok := payloads.([]ExecutePatchFanZonePayload)
	if !ok {
		return fmt.Errorf("invalid payload type: expected []ExecutePatchFanZonePayload, got %T", payloads)
	}
	
	if len(fanZonePayloads) == 0 {
		return fmt.Errorf("at least one fan zone patch payload is required")
	}

	// Validate each fan zone patch payload
	seenZones := make(map[string]bool)
	for i, fz := range fanZonePayloads {
		// Check for empty ManagerID
		if fz.ManagerID == "" {
			return fmt.Errorf("payload validation failed at payload[%d]: ManagerID is required and cannot be empty. Add a valid manager identifier", i)
		}

		// Check for empty FanZoneID
		if fz.FanZoneID == "" {
			return fmt.Errorf("payload validation failed at payload[%d]: FanZoneID is required and cannot be empty. Add a valid fan zone identifier", i)
		}

		// Check for duplicate FanZoneIDs per ManagerID
		key := fmt.Sprintf("%s:%s", fz.ManagerID, fz.FanZoneID)
		if seenZones[key] {
			return fmt.Errorf("payload validation failed: duplicate FanZoneID '%s' found for ManagerID '%s' at payload[%d]. Each fan zone can only appear once per manager. Remove duplicate entries", fz.FanZoneID, fz.ManagerID, i)
		}
		seenZones[key] = true

		// Validate payload fields - convert struct to map first
		payloadMap, err := structToMap(fz.Payload)
		if err != nil {
			return fmt.Errorf("payload validation failed at payload[%d] (ManagerID=%s, FanZoneID=%s): failed to convert payload to expected format: %w. Check field types and values match API schema", i, fz.ManagerID, fz.FanZoneID, err)
		}

		// Validate payload fields
		if errResp := utility.CheckPatchPayloadHelper(payloadMap, extendprovider.FanZoneAllowedPatchFields()); errResp != nil {
			return fmt.Errorf("payload validation failed at payload[%d] (ManagerID=%s, FanZoneID=%s): %s", i, fz.ManagerID, fz.FanZoneID, errResp.Error.Message)
		}
	}
	return nil
}

// ExecutePatchPidControllerPayload represents a PID controller patch payload
type ExecutePatchPidControllerPayload struct {
	ManagerID        string                                `json:"ManagerID"`
	PidControllerID  string                                `json:"PidControllerID"`
	Payload          extendprovider.PatchPidControllerType `json:"Payload"`
}

// ValidatePidControllerPayloads validates an array of PID controller patch payloads
func ValidatePidControllerPayloads(payloads Payload) error {
	// Type assert payload to []ExecutePatchPidControllerPayload
	pidControllerPayloads, ok := payloads.([]ExecutePatchPidControllerPayload)
	if !ok {
		return fmt.Errorf("invalid payload type: expected []ExecutePatchPidControllerPayload, got %T", payloads)
	}
	
	if len(pidControllerPayloads) == 0 {
		return fmt.Errorf("at least one PID controller patch payload is required")
	}

	// Validate each PID controller patch payload
	seenControllers := make(map[string]bool)
	for i, pc := range pidControllerPayloads {
		// Check for empty ManagerID
		if pc.ManagerID == "" {
			return fmt.Errorf("payload validation failed at payload[%d]: ManagerID is required and cannot be empty. Add a valid manager identifier", i)
		}

		// Check for empty PidControllerID
		if pc.PidControllerID == "" {
			return fmt.Errorf("payload validation failed at payload[%d]: PidControllerID is required and cannot be empty. Add a valid PID controller identifier", i)
		}

		// Check for duplicate PidControllerIDs per ManagerID
		key := fmt.Sprintf("%s:%s", pc.ManagerID, pc.PidControllerID)
		if seenControllers[key] {
			return fmt.Errorf("payload validation failed: duplicate PidControllerID '%s' found for ManagerID '%s' at payload[%d]. Each PID controller can only appear once per manager. Remove duplicate entries", pc.PidControllerID, pc.ManagerID, i)
		}
		seenControllers[key] = true

		// Validate payload fields - convert struct to map first
		payloadMap, err := structToMap(pc.Payload)
		if err != nil {
			return fmt.Errorf("payload validation failed at payload[%d] (ManagerID=%s, PidControllerID=%s): failed to convert payload to expected format: %w. Check field types and values match API schema", i, pc.ManagerID, pc.PidControllerID, err)
		}

		// Validate payload fields
		if errResp := utility.CheckPatchPayloadHelper(payloadMap, extendprovider.PidControllerAllowedPatchFields()); errResp != nil {
			return fmt.Errorf("payload validation failed at payload[%d] (ManagerID=%s, PidControllerID=%s): %s", i, pc.ManagerID, pc.PidControllerID, errResp.Error.Message)
		}
	}
	return nil
}
