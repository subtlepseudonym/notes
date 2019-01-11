package main

import (
	"fmt"
	"os"

	"github.com/subtlepseudonym/notes"

	"github.com/urfave/cli"
)

func main() {
	os.Setenv("_CLI_ZSH_AUTOCOMPLETE_HACK", "1")

	app := cli.NewApp()

	app.Usage = "write and organize notes"
	app.Description = "notes is intended to make it easy to jot down stream of consciousness notes while working in the command line and automatically back those notes up to a remote server"

	app.Version = notes.Version
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Connor Demille",
			Email: "subtlepseudonym@gmail.com",
		},
	}

	app.EnableBashCompletion = true
	app.ErrWriter = os.Stderr

	app.Commands = []cli.Command{
		ls,
		newNote,
		rm,
	}

	app.CommandNotFound = func(ctx *cli.Context, cmd string) {
		fmt.Fprintf(ctx.App.ErrWriter, "command %q not found", cmd)
		os.Exit(1)
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "runtime error: %s", err)
		os.Exit(1)
	}
}
