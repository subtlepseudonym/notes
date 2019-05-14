package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/subtlepseudonym/notes"
	"github.com/subtlepseudonym/notes/dal"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const (
	defaultLatestDepth = 5 // default number of IDs to search for latest note
)

var edit = cli.Command{
	Name:        "edit",
	ShortName:   "e",
	Usage:       "edit an existing note",
	Description: "Open a note for editing, as specified by the <noteID> argument. If no argument is provided, notes will open the most recently created note for editing",
	ArgsUsage:   "[<noteID>]",
	Action:      editAction,
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

func getNoteID(meta *notes.Meta, arg string, searchDepth int) (int, error) {
	var noteID int
	if arg != "" {
		noteID64, err := strconv.ParseInt(arg, 16, 64)
		if err != nil {
			return 0, errors.Wrap(err, "parse base 16 noteID argument failed")
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

func editAction(ctx *cli.Context) error {
	dal, err := dalpkg.NewDefaultDAL(Version) // FIXME: add option for different dal
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "initialize dal failed"), 1)
	}

	meta, err := dal.GetMeta()
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get meta failed"), 1)
	}

	noteID, err := getNoteID(meta, ctx.Args().First(), ctx.Int("latest-depth"))
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get note ID failed"), 1)
	}

	note, err := dal.GetNote(noteID)
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get note failed"), 1)
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

	file, err := ioutil.TempFile("", "note")
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "create temp file failed"), 1)
	}
	defer file.Close()

	stopChan := make(chan struct{})
	go func() {
		err := dalpkg.WatchAndUpdate(dal, note, file.Name(), ctx.Duration("update-period"), stopChan)
		if err != nil {
			// FIXME: do something with this error
		}
	}()

	body, err := notes.GetNoteBodyFromUser(file, ctx.String("editor"), note.Body)
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get note body from user failed"), 1)
	}

	stopChan <- struct{}{}
	if note.Body != body {
		note.Body = body
		changed = true
	}

	if !changed {
		return nil
	}

	// TODO: add option to not append edit to history
	note, err = note.AppendEdit(time.Now())
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "append edit to note history failed"), 1)
	}

	err = dal.SaveNote(note)
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "save note failed"), 1)
	}

	metaSize, err := meta.ApproxSize()
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get meta size failed"), 1)
	}

	meta.Size = metaSize
	meta.Notes[note.Meta.ID] = note.Meta
	err = dal.SaveMeta(meta)
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "save meta failed"), 1)
	}

	return nil
}
