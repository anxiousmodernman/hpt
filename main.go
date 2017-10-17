package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"
)

// Version can be set by build flags.
var Version = "0.0.1"

var (
	confPath = flag.String("conf", "", "path to config file")
)

type Path string

// automation for small infrastructures
func main() {

	app := cli.NewApp()
	app.Version = Version

	confFlag := cli.StringFlag{
		Name:  "conf",
		Usage: "path to config",
	}

	// a command
	app.Commands = []cli.Command{
		cli.Command{
			Name:  "plan",
			Flags: []cli.Flag{confFlag},
			Action: func(ctx *cli.Context) error {
				args := ctx.Args()
				fmt.Printf("first arg: %s", args.First())
				fmt.Printf("remaining args: %v", args.Tail())

				return nil
			},
		},
	}

	// default action
	app.Action = func(ctx *cli.Context) error {
		args := ctx.Args()

		if !args.Present() {
			return errors.New("you must provide a config file")
		}

		var paths = []string{args.First()}
		paths = append(paths, args.Tail()...)

		return run(paths...)
	}

	app.Run(os.Args)

}

// run takes a sequence of paths to config files.
func run(paths ...string) error {

	// TODO support merging multiple configs. For now, take the first.
	path := paths[0]

	conf, err := NewConfig(path)
	if err != nil {
		return err
	}

	for _, u := range conf.Users {
		_, err := ApplyUser(u, conf) // ugly
		if err != nil {
			log.Printf("ERROR: ApplyUser %s: %v", u.Name, err)
			continue
		}
		fmt.Println("hpt: created user", u.Name)
	}

	return nil
}

// plan takes a sequence of paths to config files.
func plan(paths ...string) error {
	return nil
}
