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
	if !ctx.Args().Present() {
		return cli.NewExitError(errors.New("note ID argument is required"), 1)
	}
	noteID, err := strconv.ParseInt(ctx.Args().First(), 16, 64)
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "parse base 16 noteID argument failed"), 1)
	}

	editor := defaultEditor
	if ctx.String("editor") != "" {
		editor = ctx.String("editor")
	} else if os.Getenv("EDITOR") != "" {
		editor = os.Getenv("EDITOR")
	}

	note, err := files.GetNote(int(noteID))
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get note failed"), 1)
	}

	body, err := files.GetNoteBodyFromUser(editor, note.Body)
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get note body from user failed"), 1)
	}

	options := notes.EditOptions{
		Title: ctx.String("title"),
	}

	err = notes.EditNote(int(noteID), body, options)
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "edit existing note failed"), 1)
	}

	return nil
}
