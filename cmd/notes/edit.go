package main

import (
	"strconv"
	"time"

	"github.com/subtlepseudonym/notes"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const updateHistoryLimit = 16

var edit = cli.Command{
	Name:      "edit",
	ShortName: "e",
	Usage:     "edit an existing note",
	ArgsUsage: "noteID",
	Action:    editAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "title, t",
			Usage: "note title",
		},
		cli.StringFlag{
			Name:   "editor",
			Usage:  "text editor command",
			Value:  defaultEditor,
			EnvVar: "EDTIOR",
		},
	},
}

func editAction(ctx *cli.Context) error {
	dal, err := notes.NewDefaultDAL(Version) // FIXME: add option for different dal
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "initialize dal failed"), 1)
	}

	meta, err := dal.GetMeta()
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get meta failed"), 1)
	}

	var noteID int64
	if ctx.Args().First() != "" {
		noteID, err = strconv.ParseInt(ctx.Args().First(), 16, 64)
		if err != nil {
			return cli.NewExitError(errors.Wrap(err, "parse base 16 noteID argument failed"), 1)
		}
	} else {
		noteID = int64(meta.LatestID)
	}

	note, err := dal.GetNote(int(noteID))
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get note failed"), 1)
	}

	body, err := notes.GetNoteBodyFromUser(ctx.String("editor"), note.Body)
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get note body from user failed"), 1)
	}

	var changed bool
	if note.Body != body {
		note.Body = body
		changed = true
	}

	if !time.Unix(0, 0).Equal(note.Meta.Deleted.Time) {
		note.Meta.Deleted.Time = time.Unix(0, 0) // restore soft deleted notes
		changed = true
	}

	if ctx.String("title") != "" {
		note.Meta.Title = ctx.String("title")
		changed = true
	}

	if changed {
		note.Meta.Updated = append([]notes.JSONTime{{time.Now()}}, note.Meta.Updated...)
		if len(note.Meta.Updated) > updateHistoryLimit {
			note.Meta.Updated = note.Meta.Updated[:updateHistoryLimit]
		}

		err = dal.SaveNote(note)
		if err != nil {
			return cli.NewExitError(errors.Wrap(err, "save note failed"), 1)
		}

		meta.Notes[note.Meta.ID] = note.Meta
		err = dal.SaveMeta(meta)
		if err != nil {
			return cli.NewExitError(errors.Wrap(err, "save meta failed"), 1)
		}
	}

	return nil
}
