package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/subtlepseudonym/notes"
	dalpkg "github.com/subtlepseudonym/notes/dal"

	"github.com/urfave/cli"
)

const infoDelimiter = "|"

func buildInfoCommand(dal dalpkg.DAL, meta *notes.Meta) cli.Command {
	return cli.Command{
		Name:        "info",
		Usage:       "print info",
		Description: "This command gets information about the app binary, the meta file, or specific note files and prints it in a human-friendly format. These are specified by providing no arguments, the \"meta\" argument, or a noteID respectively",
		ArgsUsage:   "[meta | <noteID>]",
		Action: func(ctx *cli.Context) error {
			return infoAction(ctx, dal, meta)
		},
	}
}

func infoAction(ctx *cli.Context, dal dalpkg.DAL, meta *notes.Meta) error {
	if !ctx.Args().Present() {
		return printAppInfo(ctx)
	}

	if ctx.Args().First() == "meta" {
		return printMetaInfo(ctx, meta)
	}

	noteID, err := strconv.ParseInt(ctx.Args().First(), 16, 64)
	if err != nil {
		return cli.NewExitError(fmt.Errorf("parse noteID argument: %w", err), 1)
	}

	note, err := dal.GetNote(int(noteID))
	if err != nil {
		return cli.NewExitError(fmt.Errorf("get note file: %w", err), 1)
	}

	return printNoteInfo(ctx, meta, note)
}

func printRows(ctx *cli.Context, rows [][]string) {
	var labelWidth int
	for _, row := range rows {
		if len(row[0]) > labelWidth {
			labelWidth = len(row[0])
		}
	}

	for _, row := range rows {
		labelPad := labelWidth - utf8.RuneCountInString(row[0])
		fmt.Fprintf(ctx.App.Writer, "%s%s %s %s\n", row[0], strings.Repeat(" ", labelPad), infoDelimiter, row[1])
	}
}

func printAppInfo(ctx *cli.Context) error {
	app := ctx.App

	rows := [][]string{
		{app.Name, app.Version},
		{"compiled", app.Compiled.Format(time.RFC3339)},
	}

	rows = append(rows, []string{"authors", app.Authors[0].String()})
	for i := 1; i < len(app.Authors); i++ {
		rows = append(rows, []string{"", app.Authors[i].String()})
	}

	for k, v := range app.ExtraInfo() {
		rows = append(rows, []string{k, v})
	}

	printRows(ctx, rows)
	return nil
}

func printMetaInfo(ctx *cli.Context, meta *notes.Meta) error {
	rows := [][]string{
		{"version", meta.Version},
		{"latest ID", strconv.Itoa(meta.LatestID)},
		{"size", strconv.Itoa(meta.Size)},
	}

	printRows(ctx, rows)
	return nil
}

func printNoteInfo(ctx *cli.Context, meta *notes.Meta, note *notes.Note) error {
	rows := [][]string{
		{"id", strconv.Itoa(note.Meta.ID)},
		{"title", note.Meta.Title},
		{"created", note.Meta.Created.Format(time.RFC3339)},
	}

	if !note.Meta.Deleted.Equal(time.Unix(0, 0)) {
		rows = append(rows, []string{"deleted", note.Meta.Deleted.Format(time.RFC3339)})
	}

	if note.Meta.History != nil {
		rows = append(rows, []string{"history", fmt.Sprintf("%s @ %d bytes", note.Meta.History[0].Updated.Format(time.RFC3339), note.Meta.History[0].Size)})
		for i := 1; i < len(note.Meta.History); i++ {
			rows = append(rows, []string{"", fmt.Sprintf("%s @ %d bytes", note.Meta.History[i].Updated.Format(time.RFC3339), note.Meta.History[i].Size)})
		}
	}

	printRows(ctx, rows)
	return nil
}
