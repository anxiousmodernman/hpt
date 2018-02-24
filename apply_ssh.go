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

// ApplySSH implements the `apply-ssh` command. As a separate command, it only
// returns an error and prints to the console internally.
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
			if len(v.IPs) < 1 {
				return errors.Errorf("expected string array ips for target: %s", target)
			}
			// TODO We only take the first ip for now.
			targetIP = v.IPs[0]
		}
	}
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
	confPath := filepath.Join("/tmp", confName)
	fmt.Println("confPath", confPath)
	cmd := fmt.Sprintf("sudo mv %s /tmp", confName)
	if _, err := ssh.Run(cmd); err != nil {
		return errors.Wrap(err, "scp failed")
	}
	// TODO get rid of this hardcoding
	cmd = fmt.Sprintf("sudo DO_ACCESS_KEY=%s DO_SECRET_ACCESS_KEY=%s hpt %s",
		os.Getenv("DO_ACCESS_KEY"), os.Getenv("DO_SECRET_ACCESS_KEY"), confPath)
	// could we do json over this line-oriented stream?
	output, done, err := ssh.Stream(cmd)
	if err != nil {
		return errors.Wrap(err, "hpt error")
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

	cleanup := fmt.Sprintf("sudo rm %s", confPath)
	if _, err := ssh.Run(cleanup); err != nil {
		return errors.Wrap(err, "rm failed")
	}
	return nil
}
