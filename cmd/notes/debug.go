// +build debug

package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/urfave/cli"
)

func (a *App) buildDebugCommand() cli.Command {
	return cli.Command{
		Name:        "debug",
		Usage:       "access debugging tools",
		Description: "Access lower-level structures, implementation details, and other debugging utilities. The behavior of this command and its subcommands are subject to breaking changes across non-major releases",
		Subcommands: []cli.Command{
			a.buildGetNote(),
			a.buildGetMeta(),
		},
	}
}

func (a *App) buildGetNote() cli.Command {
	return cli.Command{
		Name:        "get-note",
		Usage:       "print note structure",
		Description: "Print the contents of a note file as a json object",
		ArgsUsage:   "<noteID>",
		Action:      a.getNoteAction,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "no-body",
				Usage: "don't include note body",
			},
		},
	}
}

func (a *App) getNoteAction(ctx *cli.Context) error {
	if !ctx.Args().Present() {
		return fmt.Errorf("usage: noteID argument required")
	}
	n, err := strconv.ParseInt(ctx.Args().First(), 16, 64)
	if err != nil {
		return fmt.Errorf("parse noteID argument: %w", err)
	}
	noteID := int(n)

	note, err := a.dal.GetNote(noteID)
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

	_, err = fmt.Fprintln(ctx.App.Writer, string(b))
	if err != nil {
		return fmt.Errorf("write to app writer: %w", err)
	}
	return nil
}

func (a *App) buildGetMeta() cli.Command {
	return cli.Command{
		Name:        "get-meta",
		Usage:       "print meta structure",
		Description: "Print the contents of the meta file as a json object",
		Action:      a.getMetaAction,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "in-memory",
				Usage: "get meta that's currently in memory",
			},
		},
	}
}

func (a *App) getMetaAction(ctx *cli.Context) error {
	meta := a.meta
	if !ctx.Bool("in-memory") {
		var err error
		meta, err = a.dal.GetMeta()
		if err != nil {
			return fmt.Errorf("get meta from dal: %w", err)
		}
	}

	b, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal note: %w", err)
	}

	_, err = fmt.Fprintln(ctx.App.Writer, string(b))
	if err != nil {
		return fmt.Errorf("write to app writer: %w", err)
	}
	return nil
}
