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

	note, err := a.dal.GetNote(noteID)
	if err != nil {
		return fmt.Errorf("get note: %w", err)
	}

	if ctx.Bool("hard") {
		err = a.dal.RemoveNote(noteID)
		if err != nil {
			return fmt.Errorf("remove note file: %w", err)
		}

		delete(a.index, note.Meta.ID)
	} else {
		note.Meta.Deleted.Time = time.Now()
		err = a.dal.SaveNote(note)
		if err != nil {
			return fmt.Errorf("save note: %w", err)
		}
		logger.Info("note updated", zap.Int("noteID", note.Meta.ID))

		a.index[note.Meta.ID] = note.Meta
	}

	err = a.dal.SaveIndex(a.index)
	if err != nil {
		return fmt.Errorf("save index: %w", err)
	}
	logger.Info("index updated", zap.Int("length", len(a.index)))

	return nil
}
