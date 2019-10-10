package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/subtlepseudonym/notes/dal"
	"github.com/subtlepseudonym/notes/log"

	"github.com/chzyer/readline"
	"github.com/kballard/go-shellquote"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Set at compile time
var (
	Version       = "v0.0.0"
	Revision      = "git_revision"
	inInteractive = false
)

const (
	defaultEditor          = "vim"
	defaultUpdatePeriod    = 5 * time.Minute
	defaultNotesDirectory  = ".notes"
	defaultHistoryFilePath = ".nts_history"
	defaultLogFilePath     = ".nts_log"
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

	app.Before = mainBefore
	app.Action = mainAction

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "silent",
			Usage: "Prevent all logging",
		},
		cli.IntFlag{
			Name:  "verbosity",
			Usage: "Set the logging level",
			Value: int(zapcore.InfoLevel),
		},
	}

	dal, err := dalpkg.NewLocalDAL(defaultNotesDirectory, Version) // FIXME: option to use different dal
	if err != nil {
		fmt.Fprintf(os.Stderr, "runtime error: %s", errors.Wrap(err, "initialize dal failed"))
		os.Exit(1)
	}

	meta, err := dal.GetMeta()
	if err != nil {
		fmt.Fprintf(os.Stderr, "runtime error: %s", errors.Wrap(err, "get meta failed"))
		os.Exit(1)
	}

	app.Commands = []cli.Command{
		buildListCommand(dal, meta),
		buildNewCommand(dal, meta),
		buildRemoveCommand(dal, meta),
		buildEditCommand(dal, meta),
		buildInfoCommand(dal, meta),
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

	err = app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "runtime error: %s", err)
		os.Exit(1)
	}
}

func mainBefore(ctx *cli.Context) error {
	if ctx.GlobalBool("silent") {
		return nil
	}

	home, err := homedir.Dir()
	if err != nil {
		return errors.Wrap(err, "get home directory failed")
	}

	logLevel := ctx.GlobalInt("verbosity")
	logFile, err := os.OpenFile(path.Join(home, defaultNotesDirectory, defaultLogFilePath), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrap(err, "open log file failed")
	}

	logger := log.NewLogger(logFile, logLevel).With(zap.String("version", Version))
	zap.ReplaceGlobals(logger)

	return nil
}

func mainAction(ctx *cli.Context) error {
	if inInteractive {
		return nil
	} else {
		inInteractive = true
	}

	home, err := homedir.Dir()
	if err != nil {
		return errors.Wrap(err, "get home directory failed")
	}
	historyFile, err := os.OpenFile(path.Join(home, defaultNotesDirectory, defaultHistoryFilePath), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "open history file failed"), 1)
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
