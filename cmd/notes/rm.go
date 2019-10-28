package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/subtlepseudonym/notes"
	dalpkg "github.com/subtlepseudonym/notes/dal"

	"github.com/urfave/cli"
	"go.uber.org/zap"
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
	logger := zap.L().Named(ctx.Command.Name)

	if !ctx.Args().Present() {
		return cli.NewExitError(fmt.Errorf("usage: noteID argument is required"), 1)
	}
	n, err := strconv.ParseInt(ctx.Args().First(), 16, 64)
	if err != nil {
		return cli.NewExitError(fmt.Errorf("parse noteID argument: %w", err), 1)
	}
	noteID := int(n)

	note, err := dal.GetNote(noteID)
	if err != nil {
		return cli.NewExitError(fmt.Errorf("get note: %w", err), 1)
	}

	if ctx.Bool("hard") {
		err = dal.RemoveNote(noteID)
		if err != nil {
			return cli.NewExitError(fmt.Errorf("remove note file: %w", err), 1)
		}

		delete(meta.Notes, note.Meta.ID)
	} else {
		note.Meta.Deleted.Time = time.Now()
		err = dal.SaveNote(note)
		if err != nil {
			return cli.NewExitError(fmt.Errorf("save note: %w", err), 1)
		}
		logger.Info("note updated", zap.Int("noteID", note.Meta.ID))

		meta.Notes[note.Meta.ID] = note.Meta
	}

	err = dal.SaveMeta(meta)
	if err != nil {
		return cli.NewExitError(fmt.Errorf("save meta: %w", err), 1)
	}

	return nil
}
