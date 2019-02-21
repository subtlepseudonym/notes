package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/subtlepseudonym/notes"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const (
	defaultListSize            = 10
	defaultListTimeFormat      = time.RFC3339
	defaultListColumnDelimiter = " | "
)

var ls = cli.Command{
	Name:   "ls",
	Usage:  "list note info",
	Action: lsAction,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "all, a",
			Usage: "show all notes",
		},
		cli.BoolFlag{
			Name:  "long, l",
			Usage: "long format",
		},
		cli.BoolFlag{
			Name:  "deleted, d",
			Usage: "show soft deleted notes",
		},
		cli.IntFlag{
			Name:  "num, n",
			Usage: "number of notes to display",
			Value: defaultListSize,
		},
		cli.StringFlag{
			Name:  "time-format",
			Usage: "format to display timestamps in",
			Value: defaultListTimeFormat,
		},
		cli.StringFlag{
			Name:  "delimiter",
			Usage: "list column delimiter",
			Value: defaultListColumnDelimiter,
		},
	},
	UseShortOptionHandling: true,
}

func lsAction(ctx *cli.Context) error {
	dal, err := notes.NewDefaultDAL(Version) // FIXME: option to use different dal
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "initialize dal failed").Error(), 1)
	}

	meta, err := dal.GetMeta()
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get meta failed"), 1)
	}

	limit := defaultListSize
	if ctx.Bool("all") {
		limit = len(meta.Notes)
	} else if ctx.Int("num") != 0 {
		limit = ctx.Int("num")
	}

	padAmount := int(math.Log(float64(len(meta.Notes)))/math.Log(16.0) + 1.0)
	idFormat := fmt.Sprintf(" %%%dx", padAmount)

	var listed int
	idx := meta.LatestID

	for listed < limit && idx >= 0 {
		note, exists := meta.Notes[idx]
		idx--
		if !exists {
			continue
		}

		var fields []string
		fields = append(fields, fmt.Sprintf(idFormat, note.ID))

		if ctx.Bool("deleted") {
			if time.Unix(0, 0).Equal(note.Deleted.Time) {
				fields = append(fields, " ")
			} else {
				fields = append(fields, "d")
			}
		} else if !time.Unix(0, 0).Equal(note.Deleted.Time) {
			continue
		}

		if ctx.Bool("long") {
			timeFormat := defaultListTimeFormat
			if ctx.String("time-format") != "" {
				timeFormat = ctx.String("time-format")
			}
			fields = append(fields, note.Created.UTC().Format(timeFormat))
		}

		fields = append(fields, note.Title)

		fmt.Fprintln(ctx.App.Writer, strings.Join(fields, ctx.String("delimiter")))
		listed++
	}

	return nil
}
