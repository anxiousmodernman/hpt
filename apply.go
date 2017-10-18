package main

import "io"

// State represents the ultimate state of an apply call.
type State int

// possible states
const (
	Unchanged State = 0
	Changed   State = 1
)

// ApplyState is the return type of all our applies.
type ApplyState struct {
	Output  io.Reader
	Err     error
	Outcome State
}

// Error lets us concisely set an error on our ApplyState as a tail call in
// one of our apply functions.
func (a *ApplyState) Error(err error) *ApplyState {
	a.Err = err
	return a
}
