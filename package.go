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

func ApplyInstallPackages(conf Config, packages []string) *ApplyState {
	// HACK: we're doing "multiple" installs here. Really we want to flatten
	// this packages = [ "foo", "baz", "zaz" ] syntax into the higher level
	// config's Packages slice. Here we represent many states with one, and
	// return early on any error. Sad!
	var state *ApplyState
	for _, pkg := range packages {
		p := Package{pkg, Installed}
		state = ApplyPackage(conf, p)
		if state.Err != nil {
			return state.Error(
				fmt.Errorf("error installing %s: %v", pkg, state.Err))
		}
	}
	return state
}

func ApplyPackage(conf Config, p Package) *ApplyState {

	var state ApplyState
	output := bytes.NewBuffer([]byte("pkg: " + p.Name + "\n"))
	// pointer dance; we should probably just make the field a bytes.Buffer
	state.Output = output

	switch p.State {
	case Installed:
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
