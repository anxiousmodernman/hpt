package main

// Service ...
type Service struct {
	Name    string
	Status  string
	Enabled bool
}

// ApplyService ...
func ApplyService(conf Config, svc Service) *ApplyState {

	statusMap := map[string]string{
		"restarted": "restart",
		"started":   "start",
		"stopped":   "stop",
	}
	state := NewApplyState("service")

	out, err := ExecCommand("systemctl", statusMap[svc.Status], svc.Name)
	if err != nil {
		return state.Error(err)
	}
	state.Output.Write(out)

	return state
}
