package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/anxiousmodernman/easyssh"
	"github.com/pkg/errors"
	sshlib "golang.org/x/crypto/ssh"
)

func ApplySSH(target, conf, user, sshKey string) error {

	var mconf ManagerConfig
	// TODO this is duplicated in manage.go
	_, err := toml.DecodeFile(os.ExpandEnv("$HOME/.config/hpt/hpt.toml"), &mconf)
	if err != nil {
		return errors.Wrap(err, "could not open hpt.toml")
	}
	var targetIP string
	for k, v := range mconf.Hosts {
		if k == target {
			fmt.Println("config key", k)
			if len(v.IPs) < 1 {
				return errors.Errorf("expected string array ips for target: %s", target)
			}
			// TODO We only take the first ip for now.
			targetIP = v.IPs[0]
		}
	}
	// we expect hpt to be installed
	fmt.Println("target", targetIP)
	fmt.Println("config", conf)
	_, confName := filepath.Split(conf)
	fmt.Println("confName", confName)

	ssh := &easyssh.MakeConfig{
		User: user, Server: targetIP, Key: sshKey,
		HostKeyCallback: sshlib.InsecureIgnoreHostKey(),
	}
	fmt.Println("scp config to target...")
	if err := ssh.Scp(conf); err != nil {
		return errors.Wrap(err, "scp failed")
	}
	confPath := filepath.Join("/tmp", confName)
	fmt.Println("confPath", confPath)
	cmd := fmt.Sprintf("sudo mv %s /tmp", confName)
	if _, err := ssh.Run(cmd); err != nil {
		return errors.Wrap(err, "scp failed")
	}
	fmt.Println("execute hpt with config")
	cmd = fmt.Sprintf("sudo hpt %s", confPath)
	output, done, err := ssh.Stream(cmd)
	if err != nil {
		return errors.Wrap(err, "chmod +x failed")
	}

	go func() {
		for {
			select {
			case <-done:
				return
			case s := <-output:
				fmt.Println(s)
			}
		}
	}()

	fmt.Println("clean up")
	cleanup := fmt.Sprintf("sudo rm %s", confPath)
	if _, err := ssh.Run(cleanup); err != nil {
		return errors.Wrap(err, "chmod +x failed")
	}
	return nil
}
