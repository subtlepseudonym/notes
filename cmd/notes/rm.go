package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/urfave/cli"
	"go.uber.org/zap"
)

func (a *App) buildRemoveCommand() cli.Command {
	return cli.Command{
		Name:      "rm",
		Usage:     "remove an existing note",
		ArgsUsage: "<noteID>",
		Action:    a.rmAction,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "hard",
				Usage: "hard delete",
			},
		},
	}
}

func (a *App) rmAction(ctx *cli.Context) error {
	logger := a.logger.Named(ctx.Command.Name)

	if !ctx.Args().Present() {
		return fmt.Errorf("usage: noteID argument required")
	}
	n, err := strconv.ParseInt(ctx.Args().First(), 16, 64)
	if err != nil {
		return fmt.Errorf("parse noteID argument: %w", err)
	}
	noteID := int(n)

	note, err := a.data.GetNote(noteID)
	if err != nil {
		return fmt.Errorf("get note: %w", err)
	}

	if ctx.Bool("hard") {
		err = a.data.RemoveNote(noteID)
		if err != nil {
			return fmt.Errorf("remove note file: %w", err)
		}
	} else {
		note.Meta.Deleted.Time = time.Now()
		err = a.data.SaveNote(note)
		if err != nil {
			return fmt.Errorf("save note: %w", err)
		}
		logger.Info("note updated", zap.Int("noteID", note.Meta.ID))
	}

	return nil
}
