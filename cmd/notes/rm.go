package main

import (
	"strconv"

	"github.com/subtlepseudonym/notes"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var rm = cli.Command{
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
	if !ctx.Args().Present() {
		return cli.NewExitError(errors.New("note ID argument is required"), 1)
	}
	noteID, err := strconv.ParseInt(ctx.Args().First(), 16, 64)
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "parse base 16 noteID argument failed"), 1)
	}

	options := notes.RemoveOptions{
		Hard: ctx.Bool("hard"),
	}

	err = notes.RemoveNote(int(noteID), options)
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "remove note failed"), 1)
	}

	return nil
}
