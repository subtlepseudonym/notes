package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/subtlepseudonym/notes/files"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const (
	defaultListLimit      = 10
	defaultListTimeFormat = time.RFC822 // FIXME: should be alterable by config / flag
	defaultListDelimiter  = "|"         // FIXME: same here
)

var ls = cli.Command{
	Name:      "ls",
	Usage:     "list note info",
	ArgsUsage: "ls [flags]",
	Action:    lsAction,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "all, a",
			Usage: "show all notes",
		},
		cli.BoolFlag{
			Name:  "created, c",
			Usage: "show when notes where created",
		},
		cli.BoolFlag{
			Name:  "deleted, d",
			Usage: "show soft deleted notes",
		},
	},
}

func lsAction(ctx *cli.Context) error {
	meta, err := files.GetMeta()
	if os.IsNotExist(errors.Cause(err)) {
		meta, err = files.BuildNewMeta()
		if err != nil {
			return cli.NewExitError(errors.Wrap(err, "build new meta failed").Error(), 1)
		}
	} else if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get meta failed").Error(), 1)
	}

	limit := defaultListLimit
	if ctx.Bool("all") {
		limit = len(meta.Notes)
	}

	idFormat := fmt.Sprintf("%% %dx", len(meta.Notes)+1)

	var fields []string
	var listed int
	idx := len(meta.Notes) - 1

	for listed < limit && idx >= 0 {
		note := meta.Notes[idx]

		fields = append(fields, fmt.Sprintf(idFormat, note.ID))

		if ctx.Bool("all") || ctx.Bool("deleted") {
			if time.Unix(0, 0).UTC().Equal(note.Deleted.UTC()) {
				fields = append(fields, " ")
			} else {
				fields = append(fields, "d")
			}
		}

		if ctx.Bool("created") {
			fields = append(fields, note.Created.Format(defaultListTimeFormat))
		}

		fields = append(fields, note.Title)
		fmt.Fprintln(ctx.App.Writer, strings.Join(fields, defaultListDelimiter))
	}

	return nil
}
