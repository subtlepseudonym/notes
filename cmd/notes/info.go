package main

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/urfave/cli"
)

const infoDelimiter = "|"

var info = cli.Command{
	Name:        "info",
	Usage:       "print info",
	Description: "use this command to get information about the cli tool itself, the meta file, or specific note files",
	Action:      infoAction,
}

func infoAction(ctx *cli.Context) error {
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

	var labelWidth int
	for _, row := range rows {
		if len(row[0]) > labelWidth {
			labelWidth = len(row[0])
		}
	}

	for _, row := range rows {
		labelPad := labelWidth - utf8.RuneCountInString(row[0])
		fmt.Fprintf(app.Writer, "%s%s %s %s\n", row[0], strings.Repeat(" ", labelPad), infoDelimiter, row[1])
	}

	return nil
}
