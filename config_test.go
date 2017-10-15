package main

import "testing"

func TestNewConfig(t *testing.T) {

	c, err := NewConfig("testfixtures/config.toml")

	if err != nil {
		t.Errorf("bad: %v", err)
	}

	if c.Secret != 100 {
		t.Errorf("expected %v got %v", 100, c.Secret)
	}

	if len(c.Users) != 1 {
		t.Errorf("expected 1 user")
	}

	if len(c.Users[0].Groups) != 2 {
		t.Fail()
	}

	if c.Repos["somerepo"].URL != "https://github.com/foo/configs.git" {
		t.Fail()
	}

}
