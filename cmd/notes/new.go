package main

import (
	"time"

	"github.com/subtlepseudonym/notes"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const (
	defaultDateTitleFormat   = time.RFC1123
	defaultDateTitleLocation = "Local"
)

var newNote = cli.Command{
	Name:      "new",
	ShortName: "n",
	Usage:     "create a new note",
	Action:    newAction,
	Flags: []cli.Flag{
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
	},
}

func newAction(ctx *cli.Context) error {
	dal, err := notes.NewDefaultDAL(Version) // FIXME: option to use different dal
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "initialize dal failed"), 1)
	}

	body, err := notes.GetNoteBodyFromUser(ctx.String("editor"), "")
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get body from user failed"), 1)
	}

	meta, err := dal.GetMeta()
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get meta failed"), 1)
	}
	newNoteID := meta.LatestID + 1

	var title string
	if ctx.String("title") != "" {
		title = ctx.String("title")
	} else {
		title = generateDateTitle(ctx.String("title-format"), ctx.String("title-location"))
	}

	note := &notes.Note{
		Meta: notes.NoteMeta{
			ID:      newNoteID,
			Title:   title,
			Created: time.Now(),
			Deleted: time.Unix(0, 0),
		},
		Body: body,
	}

	err = dal.SaveNote(note)
	if err != nil {
		// FIXME: persist the note somewhere if saving it fails
		return cli.NewExitError(errors.Wrap(err, "save note failed"), 1)
	}

	_, exists := meta.Notes[note.Meta.ID]
	if exists {
		return cli.NewExitError(errors.New("note ID is not unique"), 1)
	}

	meta.Notes[note.Meta.ID] = note.Meta
	meta.LatestID = note.Meta.ID
	err = dal.SaveMeta(meta)
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "save meta failed"), 1)
	}

	return nil
}

func generateDateTitle(format, location string) string {
	var loc *time.Location
	l, err := time.LoadLocation(location)
	if err != nil {
		// TODO: log error
		loc = time.UTC
	} else {
		loc = l
	}

	return time.Now().In(loc).Format(format)
}
