// Package provision orchestrates the instance provisioning flow.
package provision

// Instance state constants.
const (
	StateCreating     = "creating"
	StateProvisioning = "provisioning"
	StateBooting      = "booting"
	StateRunning      = "running"
	StateStopped      = "stopped"
	StateStopping     = "stopping"
	StateTerminated   = "terminated"
	StateError        = "error"
)

// validTransitions defines which state transitions are allowed.
// Key: current state, Value: list of valid next states.
var validTransitions = map[string][]string{
	StateCreating:     {StateProvisioning, StateError, StateStopping},
	StateProvisioning: {StateBooting, StateError, StateStopping},
	StateBooting:      {StateRunning, StateError, StateStopping},
	StateRunning:      {StateStopped, StateStopping, StateError},
	StateStopped:      {StateRunning, StateStopping, StateTerminated},
	StateStopping:     {StateTerminated, StateError},
	StateTerminated:   {},
	StateError:        {StateStopping},
}

// CanTransition returns true if transitioning from the given state to the
// target state is valid according to the instance state machine.
func CanTransition(from, to string) bool {
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// ExternalState maps an internal instance state to the customer-facing state.
// Internal states are collapsed to reduce API surface complexity.
func ExternalState(internal string) string {
	switch internal {
	case StateCreating, StateProvisioning, StateBooting:
		return "starting"
	case StateRunning:
		return "running"
	case StateStopped:
		return "stopped"
	case StateStopping:
		return "stopping"
	case StateTerminated:
		return "terminated"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}

// IsTerminal returns true if the given state is a terminal state.
// Terminal states are states from which no forward progress is expected.
func IsTerminal(state string) bool {
	return state == StateTerminated || state == StateError
}
