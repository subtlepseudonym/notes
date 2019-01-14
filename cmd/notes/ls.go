package main

import (
	"github.com/subtlepseudonym/notes"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var ls = cli.Command{
	Name:   "ls",
	Usage:  "list note info",
	Action: lsAction,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "all, a",
			Usage: "show all notes",
		},
		cli.BoolFlag{
			Name:  "long, l",
			Usage: "long format",
		},
		cli.BoolFlag{
			Name:  "deleted, d",
			Usage: "show soft deleted notes",
		},
	},
}

func lsAction(ctx *cli.Context) error {
	options := notes.ListOptions{
		ShowAll:     ctx.Bool("all"),
		LongFormat:  ctx.Bool("l"),
		ShowDeleted: ctx.Bool("deleted"),
	}

	err := notes.ListNotes(ctx.App.Writer, options)
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "ls failed").Error(), 1)
	}

	return nil
}
