package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/kballard/go-shellquote"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Set at compile time
var (
	Version       = "v0.0.0"
	Revision      = "git_revision"
	inInteractive = false
)

const (
	defaultEditor       = "vim"
	defaultUpdatePeriod = 5 * time.Minute
	historyFilePath     = ".nts_history"
)

func main() {
	os.Setenv("_CLI_ZSH_AUTOCOMPLETE_HACK", "1")

	app := cli.NewApp()

	app.Usage = "write and organize notes"
	app.Description = "notes is intended to make it easy to jot down stream of consciousness notes, maintain meta data on those notes, and to organize them for fast, easy retrieval at a later date"

	app.Version = Version
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Connor Demille",
			Email: "subtlepseudonym@gmail.com",
		},
	}

	app.EnableBashCompletion = true
	app.ErrWriter = os.Stderr

	app.Action = mainAction
	app.Commands = []cli.Command{
		ls,
		newNote,
		rm,
		edit,
		info,
	}

	app.CommandNotFound = func(ctx *cli.Context, cmd string) {
		fmt.Fprintf(ctx.App.ErrWriter, "command %q not found", cmd)
		os.Exit(1)
	}

	app.ExtraInfo = func() map[string]string {
		return map[string]string{
			"revision": Revision,
		}
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "runtime error: %s", err)
		os.Exit(1)
	}
}

func mainAction(ctx *cli.Context) error {
	if inInteractive {
		return nil
	} else {
		inInteractive = true
	}

	historyFile, err := ioutil.TempFile("", ".nts_history") // FIXME: update history file in dal periodically
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "create history temp file failed"), 1)
	}
	historyPath, err := filepath.Abs(historyFile.Name())
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "get history file path failed"), 1)
	}

	config := &readline.Config{
		Prompt:            ctx.App.Name + "> ",
		HistoryFile:       historyPath,
		HistorySearchFold: true,
		AutoComplete:      readline.NewPrefixCompleter(buildPrefixCompleter(ctx.App.Commands)),
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
	}

	reader, err := readline.NewEx(config)
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "create reader failed"), 1)
	}

	for {
		line, err := reader.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		} else if err != nil { // As of readline@2972be2 err will only ever be readline.ErrInterrupt, io.EOF, nil
			return cli.NewExitError(errors.Wrap(err, "read line failed"), 1)
		}

		if strings.TrimSpace(line) == "exit" {
			break
		}

		args, err := shellquote.Split(line)
		if err != nil {
			return cli.NewExitError(errors.Wrap(err, "shellquote split failed"), 1)
		}

		err = ctx.App.Run(append([]string{ctx.App.Name}, args...))
		if err != nil {
			return cli.NewExitError(errors.Wrap(err, "app run failed"), 1)
		}
	}

	return nil
}

func buildPrefixCompleter(cmds []cli.Command) *readline.PrefixCompleter {
	completer := &readline.PrefixCompleter{}
	for _, cmd := range cmds {
		completer.Children = append(completer.Children, readline.PcItem(cmd.Name, buildPrefixCompleter([]cli.Command(cmd.Subcommands))))
	}

	return completer
}
