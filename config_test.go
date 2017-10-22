package main

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	t.Skip()

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

}

func TestExecutionPlan(t *testing.T) {

	cases := []struct {
		name, config string
		numFns       int
	}{
		{
			name:   "test1",
			config: "testfixtures/order.toml",
			numFns: 5,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			conf, err := NewConfig(tc.config)
			if err != nil {
				t.Fail()
			}
			ep, err := NewExecutionPlan(conf)
			if err != nil {
				t.Fail()
			}
			if ep == nil {
				panic("come on")
			}
			fns := 0
			for {
				fn := ep.Next()
				if fn == nil {
					break
				}
				fns++
			}
			if fns != tc.numFns {
				t.Errorf("expected %v funcs in ep, got %v", tc.numFns, fns)
			}

		})
	}
}
