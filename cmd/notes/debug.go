// +build debug

package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/subtlepseudonym/notes"
	dalpkg "github.com/subtlepseudonym/notes/dal"

	"github.com/urfave/cli"
)

func buildDebugCommand(dal dalpkg.DAL, meta *notes.Meta) cli.Command {
	return cli.Command{
		Name:        "debug",
		Usage:       "access debugging tools",
		Description: "Access lower-level structures, implementation details, and other debugging utilities. The behavior of this command and its subcommands are subject to breaking changes across non-major releases",
		Subcommands: []cli.Command{
			buildGetNote(dal),
		},
	}
}

func buildGetNote(dal dalpkg.DAL) cli.Command {
	return cli.Command{
		Name:        "get-note",
		Usage:       "print note structure",
		Description: "Print the contents of a note file as a json object",
		ArgsUsage:   "<noteID>",
		Action: func(ctx *cli.Context) error {
			return getNoteAction(ctx, dal)
		},
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "no-body",
				Usage: "don't include note body",
			},
		},
	}
}

func getNoteAction(ctx *cli.Context, dal dalpkg.DAL) error {
	if !ctx.Args().Present() {
		return fmt.Errorf("usage: noteID argument required")
	}
	n, err := strconv.ParseInt(ctx.Args().First(), 16, 64)
	if err != nil {
		return fmt.Errorf("parse noteID argument: %w", err)
	}
	noteID := int(n)

	note, err := dal.GetNote(noteID)
	if err != nil {
		return fmt.Errorf("get note: %w", err)
	}

	if ctx.Bool("no-body") {
		note.Body = "[EXCLUDED]"
	}

	b, err := json.Marshal(note)
	if err != nil {
		return fmt.Errorf("marshal note: %w", err)
	}

	fmt.Fprintln(ctx.App.Writer, string(b))
	return nil
}
