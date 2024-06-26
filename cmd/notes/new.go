package main

import (
	"fmt"
	"time"

	"github.com/subtlepseudonym/notes"

	"github.com/urfave/cli"
	"go.uber.org/zap"
)

const (
	defaultDateTitleFormat   = time.RFC1123
	defaultDateTitleLocation = "Local"
)

func (a *App) buildNewCommand() cli.Command {
	return cli.Command{
		Name:    "new",
		Aliases: []string{"n"},
		Usage:   "create a new note",
		Action:  a.newAction,
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
				Name:  "notebook",
				Usage: "specify which notebook to use. If unspecified, will use the default notebook",
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
			cli.StringFlag{
				Name:  "title-format",
				Usage: "default title time format",
				Value: defaultDateTitleFormat,
			},
			cli.StringFlag{
				Name:  "title-location",
				Usage: "default title time location",
				Value: defaultDateTitleLocation,
			},
			cli.DurationFlag{
				Name:  "update-period",
				Usage: "automatic note update period",
				Value: defaultUpdatePeriod,
			},
		},
	}
}

func (a *App) newAction(ctx *cli.Context) error {
	notebook := a.data.GetNotebook()
	logger := a.logger.Named(notebook).Named(ctx.Command.Name)

	if ctx.String("notebook") != "" {
		defer func() {
			a.data.SetNotebook(notebook)

			meta, err := a.data.GetMeta()
			if err != nil {
				logger.Error("get meta", zap.Error(err))
				return
			}
			a.meta = meta
		}()

		err := a.data.SetNotebook(ctx.String("notebook"))
		if err != nil {
			return fmt.Errorf("set notebook: %w", err)
		}
	}

	index, err := a.data.GetAllNoteMetas()
	if err != nil {
		return fmt.Errorf("get note metas: %v", err)
	}

	meta, err := a.data.GetMeta()
	if err != nil {
		return fmt.Errorf("get meta: %v", err)
	}
	a.meta = meta

	newNoteID := a.meta.LatestID + 1
	_, exists := index[newNoteID]
	if exists {
		return fmt.Errorf("note ID: must be unique")
	}

	var title string
	if ctx.String("title") != "" {
		title = ctx.String("title")
	} else {
		title = generateDateTitle(ctx.String("title-format"), ctx.String("title-location"), logger)
	}

	note := &notes.Note{
		Meta: notes.NoteMeta{
			ID:      newNoteID,
			Title:   title,
			Created: notes.JSONTime{time.Now()},
			Deleted: notes.JSONTime{time.Unix(0, 0)},
		},
	}
	a.meta.LatestID = note.Meta.ID
	err = a.data.SaveMeta(a.meta)
	if err != nil {
		return fmt.Errorf("save meta: %w", err)
	}
	logger.Info("meta latestID updated", zap.Int("metaSize", a.meta.Size))

	body, err := a.editNote(ctx, note, logger)
	if err != nil {
		return fmt.Errorf("user handoff: %w", err)
	}
	note.Body = body

	if !ctx.Bool("no-history") {
		note, err = note.AppendEdit(time.Now())
		if err != nil {
			return fmt.Errorf("append edit to note history: %w", err)
		}
	}

	err = a.data.SaveNote(note)
	if err != nil {
		// FIXME: persist the note somewhere if saving it fails
		return fmt.Errorf("save note: %w", err)
	}
	logger.Info("note updated", zap.Int("noteID", note.Meta.ID), zap.String("notebook", a.data.GetNotebook()))

	metaSize, err := a.meta.ApproxSize()
	if err != nil {
		return fmt.Errorf("get meta size: %w", err)
	}

	a.meta.Size = metaSize
	err = a.data.SaveMeta(a.meta)
	if err != nil {
		return fmt.Errorf("save meta: %w", err)
	}
	logger.Info("meta updated", zap.Int("metaSize", a.meta.Size))

	return nil
}

func generateDateTitle(format, location string, logger *zap.Logger) string {
	var loc *time.Location
	l, err := time.LoadLocation(location)
	if err != nil {
		logger.Warn("load location failed, defaulting to UTC", zap.String("location", location), zap.String("format", format), zap.Error(err))
		loc = time.UTC
	} else {
		loc = l
	}

	return time.Now().In(loc).Format(format)
}
