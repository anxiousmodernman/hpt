package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/BurntSushi/toml"
)

// A list of the possible state declarations. We need this because the toml
// parser returns many more values intermingled with the top level declarations.
// If a value is not present here, it is not considered a state that can be
// executed.
var validStates = []string{
	"user",
	"clone",
	"service",
	"group",
	"package",
	"packages",
	"file",
	"exec",
}

// ExecutionPlan ...
type ExecutionPlan struct {
	plan   []TypeIndex
	i      int
	config Config
}

// Next iterates our plan and yields apply functions.
func (ep *ExecutionPlan) Next() func() *ApplyState {

	if ep.i > len(ep.plan)-1 {
		return nil
	}
	p := ep.plan[ep.i]
	ep.i++
	switch p.T {
	case "user":
		return func() *ApplyState {
			u := ep.config.Users[p.I]
			return ApplyUser(u, ep.config)
		}
	case "clone":
		return func() *ApplyState {
			c := ep.config.Clones[p.I]
			return ApplyClone(ep.config, c)
		}
	case "exec":
		return func() *ApplyState {
			e := ep.config.Execs[p.I]
			fmt.Println("this is e", e)
			return ApplyExec(ep.config, e)
		}
	case "service":
		return func() *ApplyState {
			s := ep.config.Services[p.I]
			return ApplyService(ep.config, s)
		}
	case "group":
		return func() *ApplyState {
			g := ep.config.Groups[p.I]
			return ApplyGroup(g, ep.config)
		}
	case "package":
		return func() *ApplyState {
			p := ep.config.Packages[p.I]
			return ApplyPackage(ep.config, p)
		}
	case "packages":
		return func() *ApplyState {
			p := ep.config.InstallPackages
			return ApplyInstallPackages(ep.config, p)
		}
	case "file":
		return func() *ApplyState {
			f := ep.config.Files[p.I]
			fmt.Println("returned plan for", f)
			return ApplyFile(ep.config, f)
		}
	default:
		log.Println("unknown execution type", p.T)
		return nil
	}

	return nil
}

type TypeIndex struct {
	I int
	T string
}

// NewExecutionPlan ...
func NewExecutionPlan(conf Config) (*ExecutionPlan, error) {
	indices := make(map[string]int)
	for _, state := range validStates {
		// we start each map entry at negative one; no states have been
		//	scanned and we assume the config array is empty
		indices[state] = -1
	}

	var ep ExecutionPlan
	ep.config = conf

	// We only need exact matches on valid states in toml. This is because we
	// don't need to preserve the order of resolvers. They're application data
	// and plain serialization into the Config is enough.
	var plan []TypeIndex
	for _, key := range filterValid(conf.keys) {
		indices[key]++
		plan = append(plan, TypeIndex{indices[key], key})
	}
	ep.plan = plan
	fmt.Println("PLAN", ep.plan)

	return &ep, nil
}

// filterValid gives us a list of top-level state declarations by filtering a
// list of every key-value pair in the toml document. The parser gives us a
// slice of Key, and Key will have a string representation. The values come to
// us in the order they appear, and this is information we want to preserve, but
// we need to filter the list for top-level declarations only.
func filterValid(l []toml.Key) []string {
	var result []string
	for _, in := range l {
		for _, v := range validStates {
			if strings.TrimSpace(in.String()) == v {
				result = append(result, in.String())
			}
		}
	}
	return result
}
