package testing

import "encoding/json"

// Spy records invocations of a function for testing purposes.
// Each Call appends the provided arguments to the internal call log.
type Spy struct {
	calls [][]interface{}
}

// NewSpy creates a new Spy with an empty call log.
func NewSpy() *Spy {
	return &Spy{
		calls: make([][]interface{}, 0),
	}
}

// Call records a function call with the given arguments.
func (s *Spy) Call(args ...interface{}) {
	s.calls = append(s.calls, args)
}

// Calls returns the total number of calls recorded.
func (s *Spy) Calls() int {
	return len(s.calls)
}

// CallCount is an alias for Calls.
func (s *Spy) CallCount() int {
	return s.Calls()
}

// LastCall returns the arguments of the most recent call, or nil if no calls have been recorded.
func (s *Spy) LastCall() []interface{} {
	if len(s.calls) == 0 {
		return nil
	}
	return s.calls[len(s.calls)-1]
}

// CalledWith returns true if any recorded call matches the expected arguments.
// Arguments are compared using JSON serialization for deep equality.
func (s *Spy) CalledWith(expected ...interface{}) bool {
	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		return false
	}
	for _, call := range s.calls {
		callJSON, err := json.Marshal(call)
		if err != nil {
			continue
		}
		if string(callJSON) == string(expectedJSON) {
			return true
		}
	}
	return false
}

// Reset clears all recorded calls.
func (s *Spy) Reset() {
	s.calls = make([][]interface{}, 0)
}

// CallsArgs returns a copy of all recorded call arguments.
func (s *Spy) CallsArgs() [][]interface{} {
	result := make([][]interface{}, len(s.calls))
	for i, call := range s.calls {
		result[i] = append([]interface{}{}, call...)
	}
	return result
}
