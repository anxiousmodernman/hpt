package main

import (
	"bytes"
)

// Service ...
type Service struct {
	Name   string
	Status string
}

// ApplyService ...
func ApplyService(conf Config, svc Service) *ApplyState {
	var state ApplyState
	state.Output = bytes.NewBuffer([]byte("service: \n"))

	switch svc.Status {
	case "restarted":
		out, err := ExecCommand("systemctl", "restart", svc.Name)
		if err != nil {
			return state.Error(err)
		}
		state.Output.Write(out)
	}

	return &state
}
