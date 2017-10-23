package main

import (
	"bytes"
	"os/exec"
)

type Service struct {
	Name   string
	Status string
}

func ApplyService(conf Config, svc Service) *ApplyState {
	var state ApplyState
	state.Output = bytes.NewBuffer([]byte("fix me"))

	if svc.Status == "restarted" {
		cmd := exec.Command("systemctl", "restart", svc.Name)
		err := cmd.Run()
		if err != nil {
			return state.Error(err)
		}
	}

	return &state
}
