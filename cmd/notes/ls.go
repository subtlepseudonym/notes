package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/subtlepseudonym/notes"
	dalpkg "github.com/subtlepseudonym/notes/dal"

	"github.com/urfave/cli"
)

const (
	defaultListSize            = 10
	defaultListTimeFormat      = time.RFC3339
	defaultListColumnDelimiter = " | "
)

func buildListCommand(dal dalpkg.DAL, meta *notes.Meta) cli.Command {
	return cli.Command{
		Name:  "ls",
		Usage: "list note info",
		Action: func(ctx *cli.Context) error {
			return lsAction(ctx, dal, meta)
		},
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
			cli.BoolFlag{
				Name:  "reverse, r",
				Usage: "list notes in reverse order",
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
}

func lsAction(ctx *cli.Context, dal dalpkg.DAL, meta *notes.Meta) error {
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

	var noteList []string
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

		noteList = append(noteList, strings.Join(fields, ctx.String("delimiter")))
		listed++
	}

	if !ctx.Bool("reverse") {
		for l, r := 0, len(noteList)-1; l < r; l, r = l+1, r-1 {
			noteList[l], noteList[r] = noteList[r], noteList[l]
		}
	}

	for _, note := range noteList {
		fmt.Fprintln(ctx.App.Writer, note)
	}

	return nil
}
