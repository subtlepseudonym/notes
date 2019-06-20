package main

import (
	"strconv"
	"time"

	"github.com/subtlepseudonym/notes"
	dalpkg "github.com/subtlepseudonym/notes/dal"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func buildRemoveCommand(dal dalpkg.DAL, meta *notes.Meta) cli.Command {
	return cli.Command{
		Name:      "rm",
		Usage:     "remove an existing note",
		ArgsUsage: "<noteID>",
		Action: func(ctx *cli.Context) error {
			return rmAction(ctx, dal, meta)
		},
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "hard",
				Usage: "hard delete",
			},
		},
	}
}

func rmAction(ctx *cli.Context, dal dalpkg.DAL, meta *notes.Meta) error {
	if !ctx.Args().Present() {
		return cli.NewExitError(errors.New("note ID argument is required"), 1)
	}
	n, err := strconv.ParseInt(ctx.Args().First(), 16, 64)
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "parse base 16 noteID argument failed"), 1)
	}
	noteID := int(n)

	note, err := dal.GetNote(noteID)
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get note failed"), 1)
	}

	if ctx.Bool("hard") {
		err = dal.RemoveNote(noteID)
		if err != nil {
			return cli.NewExitError(errors.Wrap(err, "remove note file failed"), 1)
		}

		delete(meta.Notes, note.Meta.ID)
	} else {
		note.Meta.Deleted.Time = time.Now()
		err = dal.SaveNote(note)
		if err != nil {
			return cli.NewExitError(errors.Wrap(err, "save note failed"), 1)
		}

		meta.Notes[note.Meta.ID] = note.Meta
	}

	err = dal.SaveMeta(meta)
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "save meta failed"), 1)
	}

	return nil
}
