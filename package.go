package main

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
)

type PackageState string

const (
	Installed = "installed"
	Absent    = "absent"
)

type Package struct {
	Name  string
	State PackageState
}

func ApplyPackages(conf Config) []*ApplyState {

	if len(conf.InstallPackages) > 0 {
		var ip []Package
		for _, pkg := range conf.InstallPackages {
			ip = append(ip, Package{Name: pkg, State: Installed})
		}
		// prefix packages with ip
		conf.Packages = append(ip, conf.Packages...)
	}

	var results []*ApplyState

	for _, pkg := range conf.Packages {
		state := ApplyPackage(conf, pkg)
		results = append(results, state)
	}

	return results
}

func ApplyPackage(conf Config, p Package) *ApplyState {

	var state ApplyState
	output := bytes.NewBuffer([]byte("pkg: " + p.Name + "\n"))
	// pointer dance; we should probably just make the field a bytes.Buffer
	state.Output = output

	switch p.State {
	case Installed:
		// do install
		return state.Error(redhatInstall(p.Name, output))
	}

	return state.Error(fmt.Errorf("unknown package state: %v", p.State))
}

func redhatInstall(name string, output *bytes.Buffer) error {
	cmd := exec.Command("yum", "-y", "install", name)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer stdout.Close()
	err = cmd.Start()
	if err != nil {
		return err
	}
	go io.Copy(output, stdout)
	cmd.Wait()
	return nil
}
