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
	Clones          []Clone           `toml:"clone"`
	Files           []File            `toml:"file"`
	Execs           []Exec            `toml:"exec"`
	// keys is embedded at runtime after decoding the toml file
	keys []toml.Key
}

// Repo ...
type Repo struct {
	URL      string `toml:"url"`
	Dest     string `toml:"dest"`
	User     string `toml:"user"`
	Key      string `toml:"key"`
	Password string `toml:"password"`
}

// Bucket ...
type Bucket struct {
	URL  string `toml:"url"`
	Name string `toml:"name"`
}

// NewConfig makes a config.
func NewConfig(path string) (Config, error) {
	var c Config
	md, err := toml.DecodeFile(path, &c)
	if err != nil {
		return c, err
	}
	c.keys = md.Keys()

	return c, nil
}
