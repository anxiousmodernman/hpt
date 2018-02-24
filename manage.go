package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/anxiousmodernman/easyssh"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"

	"github.com/Rudd-O/curvetls"
	sshlib "golang.org/x/crypto/ssh"
)

var _ = errors.Wrap

var managedBinary = "/bin/hpt"

type Keypair struct {
	Pub  curvetls.Pubkey
	Priv curvetls.Privkey
}

// Manage is our function to use SSH to bring a box under management. We ssh in,
// upload the hpt binary, put the binary in /usr/bin, make it executable, install
// our systemd service and socket files, bounce the systemd daemon, and open the
// firewall for our hpt manager. We also put keys at /etc/hpt/keys/.
func Manage(target, sshUser, sshPrivKeyPath string) error {

	fmt.Println("target", target)
	// Read main hpt toml config;
	var mconf ManagerConfig
	_, err := toml.DecodeFile(os.ExpandEnv("$HOME/.config/hpt/hpt.toml"), &mconf)
	if err != nil {
		return errors.Wrap(err, "could not open hpt.toml")
	}
	fmt.Println("management config:", mconf)
	// We will connect to this IP
	var ip string

	// find our target(s) in the config by name
	for k, v := range mconf.Hosts {
		if k == target {
			fmt.Println("config key", k)
			if len(v.IPs) < 1 {
				return errors.Errorf("expected string array ips for target: %s", target)
			}
			// TODO We only take the first ip for now.
			ip = v.IPs[0]
		}
	}

	// TODO check if key exists before creating a new one

	keystoreDir := filepath.Join(os.Getenv("HOME"), ".config", "hpt")
	keystorePath := filepath.Join(keystoreDir, "keys.db")
	if err := os.MkdirAll(keystoreDir, os.FileMode(0700)); err != nil {
		return errors.Wrap(err, "error creating config dir")
	}
	fmt.Println("keystorePath:", keystorePath)

	// generate a keypair
	priv, pub, err := curvetls.GenKeyPair()
	if err != nil {
		return errors.Wrap(err, "could not generate key pair")
	}
	pair := Keypair{pub, priv}

	// store in boltdb
	db, err := bolt.Open(keystorePath, 0600, nil)
	if err != nil {
		return errors.Wrap(err, "could not open keys.db")
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("keys"))
		if err != nil {
			return err
		}
		encoded, err := gobEncode(&pair)
		if err != nil {
			return err
		}
		// keyed by ip address; long term we'll want something different
		if err := b.Put([]byte(ip), encoded); err != nil {
			return err
		}
		return nil
	})

	fmt.Println("added keypair for", ip)
	err = scpBinary(sshUser, sshPrivKeyPath, ip)

	return err
}

// we need a sudo user for this
func scpBinary(user, sshKey, targetIP string) error {
	ssh := &easyssh.MakeConfig{
		User: user, Server: targetIP, Key: sshKey,
		// Insecure means we accept any host key. This used to be the default
		// behavior, but that changed: https://github.com/golang/go/issues/19767
		// Will need to configure this somehow from hpt.conf, I expect
		HostKeyCallback: sshlib.InsecureIgnoreHostKey(),
	}
	fmt.Println("scp hpt binary from this directory to target...")
	// TODO cheating here. Lookup our own binary!
	if err := ssh.Scp("hpt"); err != nil {
		return errors.Wrap(err, "scp failed")
	}
	fmt.Println("move hpt binary to", managedBinary, "...")
	cmd := fmt.Sprintf("sudo mv hpt %s", managedBinary)
	if _, err := ssh.Run(cmd); err != nil {
		return errors.Wrap(err, "scp failed")
	}
	fmt.Println("make hpt executable...")
	cmd = fmt.Sprintf("sudo chmod +x %s", managedBinary)
	if _, err := ssh.Run(cmd); err != nil {
		return errors.Wrap(err, "chmod +x failed")
	}
	return nil
}

func gobEncode(v interface{}) ([]byte, error) {
	// Note: v must be a pointer
	var buf []byte
	b := bytes.NewBuffer(buf)
	enc := gob.NewEncoder(b)

	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func gobDecode(data []byte, v interface{}) error {
	// Note: v must be a pointer
	b := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(b)
	return decoder.Decode(v)

}
