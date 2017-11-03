package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/urfave/cli"
)

// Version can be set by build flags.
var Version = "0.0.1"

var (
	confPath = flag.String("conf", "", "path to config file")
)

// automation for small infrastructures
func main() {

	app := cli.NewApp()
	app.Version = Version

	confFlag := cli.StringFlag{
		Name:  "conf",
		Usage: "path to config",
	}

	sshUser := cli.StringFlag{
		Name:  "user",
		Usage: "ssh user to connect as",
	}

	sshPrivKeyPath := cli.StringFlag{
		Name:  "sshIdent",
		Usage: "path to private ssh key",
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name:  "apply-ssh",
			Flags: []cli.Flag{confFlag, sshUser, sshPrivKeyPath},
			Usage: "run an hpt config over ssh",
			Action: func(ctx *cli.Context) error {
				if !ctx.Args().Present() {
					fmt.Println("you must provide an hpt config")
					os.Exit(1)
				}
				user, key := ctx.String("user"), ctx.String("sshIdent")
				err := ApplySSH(ctx.Args().First(), ctx.Args().Get(1), user, key)
				return err
			},
		},
		cli.Command{
			Name:  "manage",
			Flags: []cli.Flag{confFlag, sshUser, sshPrivKeyPath},
			Usage: "bring a box under management",
			Action: func(ctx *cli.Context) error {
				if !ctx.Args().Present() {
					fmt.Println("you must provide an IP to manage")
					os.Exit(1)
				}
				user, key := ctx.String("user"), ctx.String("sshIdent")
				err := Manage(ctx.Args().First(), user, key)
				return err
			},
		},
		cli.Command{
			Name:  "plan",
			Flags: []cli.Flag{confFlag},
			Action: func(ctx *cli.Context) error {
				return nil
			},
		},
		cli.Command{
			Name:  "serve",
			Flags: []cli.Flag{confFlag},
			Usage: "start an hpt daemon",
			Action: func(ctx *cli.Context) error {
				return nil
			},
		},
	}

	// default action
	app.Action = func(ctx *cli.Context) error {
		args := ctx.Args()
		if !args.Present() {
			return errors.New("you must provide a state file")
		}
		if args.First() == "ssh-apply" {
			fmt.Println("did you mean \"apply-ssh\"?")
			os.Exit(1)
		}
		var paths = []string{args.First()}
		paths = append(paths, args.Tail()...)
		return run(paths...)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

var (
	boldRed = color.New(color.FgRed).Add(color.Bold)
	red     = color.New(color.FgRed)
	white   = color.New(color.FgWhite)
	yellow  = color.New(color.FgYellow)
	blue    = color.New(color.FgBlue)
)

// run takes a sequence of paths to config files.
func run(paths ...string) error {

	// TODO support multiple configs
	path := paths[0]

	printState := func(s *ApplyState) {
		if s.Err != nil {
			red.Print(string(s.RenderShell()))
		}
		blue.Println(string(s.RenderShell()))
	}
	conf, err := NewConfig(path)
	if err != nil {
		return err
	}
	ep, err := NewExecutionPlan(conf)
	if err != nil {
		return err
	}
	for {
		fn := ep.Next()
		if fn == nil {
			break
		}
		state := fn()
		printState(state)
	}

	return nil
}

// plan takes a sequence of paths to config files.
func plan(paths ...string) error {
	return nil
}
