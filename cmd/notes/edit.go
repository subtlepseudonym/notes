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

const (
	defaultLatestDepth = 5 // default number of IDs to search for latest note
)

func buildEditCommand(dal dalpkg.DAL, meta *notes.Meta) cli.Command {
	return cli.Command{
		Name:        "edit",
		ShortName:   "e",
		Usage:       "edit an existing note",
		Description: "Open a note for editing, as specified by the <noteID> argument. If no argument is provided, notes will open the most recently created note for editing",
		ArgsUsage:   "[<noteID>]",
		Action: func(ctx *cli.Context) error {
			return editAction(ctx, dal, meta)
		},
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "no-watch",
				Usage: "don't save note in background",
			},
			cli.BoolFlag{
				Name:  "no-history",
				Usage: "don't record activity in edit history",
			},
			cli.StringFlag{
				Name:  "title, t",
				Usage: "note title",
			},
			cli.StringFlag{
				Name:   "editor",
				Usage:  "text editor command",
				Value:  defaultEditor,
				EnvVar: "EDITOR",
			},
			cli.IntFlag{
				Name:  "latest-depth",
				Usage: "number of IDs to search from latest ID to find latest note",
				Value: defaultLatestDepth,
			},
			cli.DurationFlag{
				Name:  "update-period",
				Usage: "automatic note update period",
				Value: defaultUpdatePeriod,
			},
		},
	}
}

func getNoteID(meta *notes.Meta, arg string, searchDepth int) (int, error) {
	var noteID int
	if arg != "" {
		noteID64, err := strconv.ParseInt(arg, 16, 64)
		if err != nil {
			return 0, fmt.Errorf("parse noteID argument: %w", err)
		}
		noteID = int(noteID64)
	} else {
		for i := 0; i < searchDepth; i++ {
			if _, exists := meta.Notes[meta.LatestID-i]; exists {
				noteID = meta.LatestID - i
				break
			}
		}
	}

	if noteID == 0 {
		// FIXME: may want to log a note that this is based upon content of the meta rather than a DAL existence check
		return 0, fmt.Errorf("latest note ID ⊄ [%x,%x], try using noteID argument or --latest-depth", meta.LatestID-searchDepth, meta.LatestID)
	}

	return noteID, nil
}

func editAction(ctx *cli.Context, dal dalpkg.DAL, meta *notes.Meta) error {
	logger := zap.L().Named(ctx.Command.Name)

	noteID, err := getNoteID(meta, ctx.Args().First(), ctx.Int("latest-depth"))
	if err != nil {
		return fmt.Errorf("get note ID: %w", err)
	}

	note, err := dal.GetNote(noteID)
	if err != nil {
		return fmt.Errorf("get note: %w", err)
	}

	var changed bool
	if !note.Meta.Deleted.Time.Equal(time.Unix(0, 0)) {
		note.Meta.Deleted.Time = time.Unix(0, 0) // restore soft deleted notes
		changed = true
	}

	if ctx.String("title") != "" {
		note.Meta.Title = ctx.String("title")
		changed = true
	}

	body, err := editNote(ctx, dal, meta, note)
	if err != nil {
		return fmt.Errorf("user handoff: %w", err)
	}

	if note.Body != body {
		note.Body = body
		changed = true
	}

	if !changed {
		return nil
	}

	if !ctx.Bool("no-history") {
		note, err = note.AppendEdit(time.Now())
		if err != nil {
			return fmt.Errorf("append edit to note history: %w", err)
		}
	}

	err = dal.SaveNote(note)
	if err != nil {
		return fmt.Errorf("save note: %w", err)
	}
	logger.Info("note updated", zap.Int("noteID", note.Meta.ID))

	metaSize, err := meta.ApproxSize()
	if err != nil {
		return fmt.Errorf("get meta size: %w", err)
	}

	meta.Size = metaSize
	meta.Notes[note.Meta.ID] = note.Meta
	err = dal.SaveMeta(meta)
	if err != nil {
		return fmt.Errorf("save meta: %w", err)
	}
	logger.Info("meta updated", zap.Int("metaSize", meta.Size))

	return nil
}
