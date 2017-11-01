package main

import (
	"fmt"
	"path/filepath"

	"github.com/anxiousmodernman/easyssh"
	"github.com/pkg/errors"
	sshlib "golang.org/x/crypto/ssh"
)

func ApplySSH(conf, user, privKey string) error {
	// we expect hpt to be installed
	_, confName := filepath.Split(conf)

	ssh := &easyssh.MakeConfig{
		User: user, Server: targetIP, Key: sshKey,
		HostKeyCallback: sshlib.InsecureIgnoreHostKey(),
	}
	fmt.Println("scp config to target...")
	if err := ssh.Scp(conf); err != nil {
		return errors.Wrap(err, "scp failed")
	}
	fmt.Println("move config to /tmp")
	confPath := filepath.Join("tmp", confName)
	cmd := fmt.Sprintf("sudo mv %s /tmp", confName, confPath)
	if _, err := ssh.Run(cmd); err != nil {
		return errors.Wrap(err, "scp failed")
	}
	fmt.Println("execute hpt with config")
	cmd = fmt.Sprintf("sudo hpt %s", confPath)
	if _, err := ssh.Run(cmd); err != nil {
		return errors.Wrap(err, "chmod +x failed")
	}
	fmt.Println("clean up")
	cleanup := fmt.Sprintf("sudo rm %s", confPath)
	if _, err := ssh.Run(cleanup); err != nil {
		return errors.Wrap(err, "chmod +x failed")
	}
	return nil
}
