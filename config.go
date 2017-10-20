package main

import (
	"github.com/BurntSushi/toml"
)

// Config is our container for passing any kind of applyable state.
type Config struct {
	InstallPackages []string          `toml:"packages"`
	Secret          int               `toml:"secret"`
	Groups          []Group           `toml:"group"`
	Users           []User            `toml:"user"`
	Repos           map[string]Repo   `toml:"repo"`
	Buckets         map[string]Bucket `toml:"bucket"`
	Services        []Service         `toml:"service"`
	Packages        []Package         `toml:"package"`
}

// Repo ...
type Repo struct {
	URL string `toml:"url"`
}

// Bucket ...
type Bucket struct {
	URL  string `toml:"url"`
	Name string `toml:"name"`
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
