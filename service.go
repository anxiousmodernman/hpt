package main

import (
	"bytes"
	"os/exec"
)

type Service struct {
	Name   string
	Status string
}

func ApplyServices(conf Config) []*ApplyState {
	var result []*ApplyState
	for _, s := range conf.Services {
		state := ApplyService(conf, s)
		result = append(result, state)
	}
	return result
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
