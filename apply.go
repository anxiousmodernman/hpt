package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

// State represents the ultimate state of an apply call.
type State int

// possible states
const (
	Unchanged State = 0
	Changed   State = 1
)

// ApplyState is the return type of all our applies.
type ApplyState struct {
	Name    string
	Output  *bytes.Buffer
	Err     error
	Outcome State
}

// NewApplyState is a convenience constructor for an ApplyState.
func NewApplyState(name string) *ApplyState {
	a := &ApplyState{
		Name:   name,
		Output: bytes.NewBuffer([]byte("")),
	}
	return a
}

// RenderShell lets us print the outcome of each apply.
func (a *ApplyState) RenderShell() []byte {

	/*
		+------------------------
		 NAME
		 OUTCOME
		 ATTR: FOO
		 ATTR: BAZ
		 ERROR: <the err.Error() output>
		 OUTPUT: <buffer>
	*/

	output := bytes.NewBuffer([]byte("+------------------------\n"))

	write := func(key string, value interface{}) {
		if value == nil {
			return
		}
		switch value.(type) {
		case State:
			formatted := fmt.Sprintf(" %s: %v\n", key, value)
			output.WriteString(formatted)
		case string:
			formatted := fmt.Sprintf(" %s: %s\n", key, value)
			output.WriteString(formatted)
		case error:
			formatted := fmt.Sprintf(" %s: %v\n", key, value)
			output.WriteString(formatted)
		case *bytes.Buffer:
			formatted := fmt.Sprintf(" %s: ", key)
			output.WriteString(formatted)
			buf := value.(*bytes.Buffer)
			output.Write(buf.Bytes())
		}
	}

	write("NAME", a.Name)
	write("OUTCOME", a.Outcome)
	write("ERROR", a.Err)
	write("OUTPUT", a.Output)
	return output.Bytes()
}

// Error lets us concisely set an error on our ApplyState as a tail call in
// one of our apply functions.
func (a *ApplyState) Error(err error) *ApplyState {
	// we check for errors from package exec because we need to capture the
	// output written to stderr from failed commands
	if err, ok := err.(*exec.ExitError); ok {
		if a.Output != nil {
			a.Output.Write(err.Stderr)
		}
	}
	a.Err = err
	return a
}

// Errorf is like Error but takes a format string.
func (a *ApplyState) Errorf(format string, err error) *ApplyState {
	a.Err = fmt.Errorf(format, err)
	return a
}
