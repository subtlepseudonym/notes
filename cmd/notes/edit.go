package main

import (
	"os"
	"strconv"

	"github.com/subtlepseudonym/notes"
	"github.com/subtlepseudonym/notes/files"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var edit = cli.Command{
	Name:      "edit",
	Usage:     "edit an existing note",
	ArgsUsage: "noteID",
	Action:    editAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "title, t",
			Usage: "note title",
		},
		cli.StringFlag{
			Name:  "editor",
			Usage: "text editor command",
		},
	},
}

func editAction(ctx *cli.Context) error {
	editor := defaultEditor
	if ctx.String("editor") != "" {
		editor = ctx.String("editor")
	} else if os.Getenv("EDITOR") != "" {
		editor = os.Getenv("EDITOR")
	}

	dal, err := files.NewDefaultDAL(Version) // FIXME: add option for different dal
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "initialize dal failed").Error(), 1)
	}

	meta, err := dal.GetMeta()
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get meta failed").Error(), 1)
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

	body, err := files.GetNoteBodyFromUser(editor, note.Body)
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get note body from user failed"), 1)
	}
	note.Body = body

	options := notes.EditOptions{
		Title: ctx.String("title"),
	}

	_, _, err = notes.EditNote(note, options, dal)
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "edit existing note failed"), 1)
	}

	return nil
}
