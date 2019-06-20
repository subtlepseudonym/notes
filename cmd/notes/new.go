package main

import (
	"io/ioutil"
	"time"

	"github.com/subtlepseudonym/notes"
	"github.com/subtlepseudonym/notes/dal"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"go.uber.org/zap"
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
		cli.BoolFlag{
			Name:  "no-watch",
			Usage: "don't save note in background",
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

func newAction(ctx *cli.Context) error {
	dal, err := dalpkg.NewLocalDAL(defaultNotesDirectory, Version) // FIXME: option to use different dal
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "initialize dal failed"), 1)
	}

	meta, err := dal.GetMeta()
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get meta failed"), 1)
	}
	newNoteID := meta.LatestID + 1

	_, exists := meta.Notes[newNoteID]
	if exists {
		return cli.NewExitError(errors.New("note ID is not unique"), 1)
	}

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
			Created: notes.JSONTime{time.Now()},
			Deleted: notes.JSONTime{time.Unix(0, 0)},
		},
	}

	file, err := ioutil.TempFile("", "note")
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "create temporary file failed"), 1)
	}
	defer file.Close()

	stop := make(chan struct{})
	if !ctx.Bool("no-watch") {
		go func() {
			err := dalpkg.WatchAndUpdate(dal, note, file.Name(), ctx.Duration("update-period"), stop, Logger)
			if err != nil {
				Logger.Error("watch and updated failed", zap.Error(err), zap.Int("noteID", note.Meta.ID), zap.String("filename", file.Name()))
			}
		}()
	}

	body, err := notes.GetNoteBodyFromUser(file, ctx.String("editor"), "")
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get body from user failed"), 1)
	}
	note.Body = body
	close(stop)

	// TODO: add option to not append edit to history
	note, err = note.AppendEdit(time.Now())
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "append edit to note history failed"), 1)
	}

	err = dal.SaveNote(note)
	if err != nil {
		// FIXME: persist the note somewhere if saving it fails
		return cli.NewExitError(errors.Wrap(err, "save note failed"), 1)
	}

	metaSize, err := meta.ApproxSize()
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get meta size failed"), 1)
	}

	meta.Size = metaSize
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
		Logger.Warn("load location failed, defaulting to UTC", zap.String("location", location), zap.String("format", format), zap.Error(err))
		loc = time.UTC
	} else {
		loc = l
	}

	return time.Now().In(loc).Format(format)
}
