package main

import (
	"github.com/urfave/cli"
)

var rmNote = cli.Command{
	Name:      "rm",
	Usage:     "remove an existing note",
	ArgsUsage: "rm [flags] noteID",
	Action:    rmAction,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "hard",
			Usage: "hard delete",
		},
	},
}

func rmAction(ctx *cli.Context) error {
	return nil
}
