package main

import (
	"github.com/BurntSushi/toml"
)

// Config ...
type Config struct {
	Secret  int               `toml:"secret"`
	Users   []User            `toml:"user"`
	Repos   map[string]Repo   `toml:"repo"`
	Buckets map[string]Bucket `toml:"bucket"`
}

type Repo struct {
	URL string
}

type Bucket struct {
	URL, Name string
}

// User ...
type User struct {
	Name          string   `toml:"name"`
	Home          string   `toml:"home"`
	Groups        []string `toml:"groups"`
	SSHPublicKey  string   `toml:"ssh_public_key"`
	SSHPrivateKey string   `toml:"ssh_private_key"`
}

// NewConfig makes a config.
func NewConfig(path string) (Config, error) {
	var c Config
	_, err := toml.DecodeFile(path, &c)
	if err != nil {
		return c, err
	}
	return c, nil
}


/*
ansible:
# Add the user 'james' with a bash shell, appending the group 'admins' and 'developers' to the user's groups
- user:
    name: james
    shell: /bin/bash
    groups: admins,developers
    append: yes

# Remove the user 'johnd'
- user:
    name: johnd
    state: absent
    remove: yes
*/