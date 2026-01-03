package ws

import (
	"fmt"
	"sync"
)

// StateMachine is a generic state machine with defined transitions.
type StateMachine[S comparable] struct {
	mu          sync.RWMutex
	current     S
	transitions map[S][]S
	onChange    func(from, to S)
}

// NewStateMachine creates a new state machine with initial state and valid transitions.
func NewStateMachine[S comparable](initial S, transitions map[S][]S) *StateMachine[S] {
	return &StateMachine[S]{
		current:     initial,
		transitions: transitions,
	}
}

// OnChange sets a callback for state changes.
func (sm *StateMachine[S]) OnChange(fn func(from, to S)) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.onChange = fn
}

// Current returns the current state.
func (sm *StateMachine[S]) Current() S {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.current
}

// CanTransition checks if transition to the given state is valid.
func (sm *StateMachine[S]) CanTransition(to S) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.canTransitionLocked(to)
}

func (sm *StateMachine[S]) canTransitionLocked(to S) bool {
	valid, ok := sm.transitions[sm.current]
	if !ok {
		return false
	}
	for _, s := range valid {
		if s == to {
			return true
		}
	}
	return false
}

// Transition attempts to transition to the given state.
// Returns an error if the transition is invalid.
func (sm *StateMachine[S]) Transition(to S) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.canTransitionLocked(to) {
		return &InvalidTransitionError[S]{From: sm.current, To: to}
	}

	from := sm.current
	sm.current = to

	if sm.onChange != nil {
		sm.onChange(from, to)
	}

	return nil
}

// MustTransition transitions to the given state, panicking on invalid transition.
func (sm *StateMachine[S]) MustTransition(to S) {
	if err := sm.Transition(to); err != nil {
		panic(err)
	}
}

// ForceTransition transitions to the given state without validation.
// Use sparingly, primarily for recovery scenarios.
func (sm *StateMachine[S]) ForceTransition(to S) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	from := sm.current
	sm.current = to

	if sm.onChange != nil {
		sm.onChange(from, to)
	}
}

// Set sets the current state without triggering onChange.
// Use for initialization or restoration scenarios.
func (sm *StateMachine[S]) Set(state S) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.current = state
}

// ValidNextStates returns the list of valid states to transition to.
func (sm *StateMachine[S]) ValidNextStates() []S {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.transitions[sm.current]
}

// Is checks if the current state matches any of the given states.
func (sm *StateMachine[S]) Is(states ...S) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	for _, s := range states {
		if sm.current == s {
			return true
		}
	}
	return false
}

// InvalidTransitionError is returned when an invalid state transition is attempted.
type InvalidTransitionError[S comparable] struct {
	From S
	To   S
}

func (e *InvalidTransitionError[S]) Error() string {
	return fmt.Sprintf("invalid state transition: %v -> %v", e.From, e.To)
}

// StateHistory tracks state transitions for debugging/testing.
type StateHistory[S comparable] struct {
	mu      sync.Mutex
	history []StateTransition[S]
	maxSize int
}

// StateTransition represents a single state transition.
type StateTransition[S comparable] struct {
	From S
	To   S
}

// NewStateHistory creates a new state history tracker.
func NewStateHistory[S comparable](maxSize int) *StateHistory[S] {
	return &StateHistory[S]{
		history: make([]StateTransition[S], 0, maxSize),
		maxSize: maxSize,
	}
}

// Record records a state transition.
func (h *StateHistory[S]) Record(from, to S) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.history) >= h.maxSize {
		// Remove oldest
		h.history = h.history[1:]
	}
	h.history = append(h.history, StateTransition[S]{From: from, To: to})
}

// All returns all recorded transitions.
func (h *StateHistory[S]) All() []StateTransition[S] {
	h.mu.Lock()
	defer h.mu.Unlock()
	result := make([]StateTransition[S], len(h.history))
	copy(result, h.history)
	return result
}

// Last returns the last n transitions.
func (h *StateHistory[S]) Last(n int) []StateTransition[S] {
	h.mu.Lock()
	defer h.mu.Unlock()

	if n > len(h.history) {
		n = len(h.history)
	}
	start := len(h.history) - n
	result := make([]StateTransition[S], n)
	copy(result, h.history[start:])
	return result
}

// Clear clears all recorded transitions.
func (h *StateHistory[S]) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.history = h.history[:0]
}

// TrackingStateMachine wraps a StateMachine with history tracking.
func TrackingStateMachine[S comparable](sm *StateMachine[S], maxHistory int) (*StateMachine[S], *StateHistory[S]) {
	history := NewStateHistory[S](maxHistory)
	originalOnChange := sm.onChange

	sm.OnChange(func(from, to S) {
		history.Record(from, to)
		if originalOnChange != nil {
			originalOnChange(from, to)
		}
	})

	return sm, history
}
