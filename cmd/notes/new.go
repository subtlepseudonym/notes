package main

import (
	"os"

	"github.com/subtlepseudonym/notes"
	"github.com/subtlepseudonym/notes/files"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const defaultEditor = "vim"

var newNote = cli.Command{
	Name:   "new",
	Usage:  "create a new note",
	Action: newAction,
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

func newAction(ctx *cli.Context) error {
	options := notes.NoteOptions{
		Title: ctx.String("title"),
	}

	editor := defaultEditor
	if ctx.String("editor") != "" {
		editor = ctx.String("editor")
	} else if os.Getenv("EDITOR") != "" {
		editor = os.Getenv("EDITOR")
	}

	meta, err := files.GetMeta(ctx.App.Version)
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get meta failed").Error(), 1)
	}

	body, err := files.GetNoteBodyFromUser(editor, "")
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get body from user failed").Error(), 1)
	}

	_, _, err = notes.NewNote(body, options, meta)
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "create new failed").Error(), 1)
	}

	return nil
}
