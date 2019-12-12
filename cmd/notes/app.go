package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/subtlepseudonym/notes"
	dalpkg "github.com/subtlepseudonym/notes/dal"

	"github.com/Masterminds/semver"
	"github.com/chzyer/readline"
	"github.com/kballard/go-shellquote"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Set at compile time
var (
	Version   = "v0.0.0"
	Revision  = "git_revision"
	BuildTags = "list of flags"
)

const (
	defaultNotesDirectory  = ".notes"
	defaultHistoryFilePath = ".nts_history"
	defaultLogFilePath     = ".nts_log"
)

type App struct {
	*cli.App

	homeDir string
	logger  *zap.Logger
	dal     dalpkg.DAL
	meta    *notes.Meta

	inInteractive bool
}

func New() (*App, error) {
	app := &App{
		App: cli.NewApp(),
	}

	app.Usage = "write and organize notes"
	app.Description = "notes is intended to make it easy to jot down stream of consciousness notes, maintain meta data on those notes, and to organize them for fast, easy retrieval at a later date"

	app.Version = Version
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Connor Demille",
			Email: "subtlepseudonym@gmail.com",
		},
	}

	extraInfo := make(map[string]string)
	extraInfo["revision"] = Revision
	if BuildTags != "" {
		extraInfo["tags"] = BuildTags
	}
	app.ExtraInfo = func() map[string]string {
		return extraInfo
	}

	dal, err := dalpkg.NewLocalDAL(defaultNotesDirectory, Version) // FIXME: option to use different dal
	if err != nil {
		return nil, fmt.Errorf("initialize dal: %v", err)
	}
	app.dal = dal

	meta, err := dal.GetMeta()
	if err != nil {
		return nil, fmt.Errorf("get meta: %v", err)
	}
	app.meta = meta

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

	app.Commands = []cli.Command{
		buildDebugCommand(dal, meta),
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

	app.App.Before = app.before
	app.Action = app.interactiveMode

	app.EnableBashCompletion = true
	app.Writer = os.Stdout
	app.ErrWriter = os.Stderr

	return app, nil
}

func (a *App) before(ctx *cli.Context) error {
	if a.homeDir == "" {
		home, err := homedir.Dir()
		if err != nil {
			return fmt.Errorf("get home directory: %w", err)
		}
		a.homeDir = home
	}

	if a.logger != nil {
		logger, err := a.initLogging(ctx)
		if err != nil {
			return fmt.Errorf("init logging: %v", err)
		}
		a.logger = logger
	}

	return nil
}

func (a *App) initLogging(ctx *cli.Context) (*zap.Logger, error) {
	if ctx.GlobalBool("silent") {
		return zap.NewNop(), nil
	}

	logLevel := ctx.GlobalInt("verbosity")
	logFile, err := os.OpenFile(path.Join(a.homeDir, defaultNotesDirectory, defaultLogFilePath), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.NanosDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewJSONEncoder(encoderConfig)
	core := zapcore.NewCore(encoder, logFile, zapcore.Level(int8(logLevel)))

	return zap.New(core).With(zap.String("version", a.Version)), nil
}

func (a *App) interactiveMode(ctx *cli.Context) error {
	if a.inInteractive {
		return nil
	}
	a.inInteractive = true

	historyFile, err := os.OpenFile(path.Join(a.homeDir, defaultNotesDirectory, defaultHistoryFilePath), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return cli.NewExitError(fmt.Errorf("open history file: %w", err), 1)
	}
	historyPath, err := filepath.Abs(historyFile.Name())
	if err != nil {
		return cli.NewExitError(fmt.Errorf("get history file path: %w", err), 1)
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
		return cli.NewExitError(fmt.Errorf("create reader: %w", err), 1)
	}

	for {
		line, err := reader.Readline()
		if errors.Is(err, readline.ErrInterrupt) {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if errors.Is(err, io.EOF) {
			break
		} else if err != nil { // As of readline@2972be2 err will only ever be readline.ErrInterrupt, io.EOF, nil
			return cli.NewExitError(fmt.Errorf("read line: %w", err), 1)
		}

		if strings.TrimSpace(line) == "exit" {
			break
		}

		args, err := shellquote.Split(line)
		if err != nil {
			return cli.NewExitError(fmt.Errorf("shellquote split: %w", err), 1)
		}

		err = ctx.App.Run(append([]string{ctx.App.Name}, args...))
		if err != nil {
			if errors.Is(err, &cli.ExitError{}) {
				return cli.NewExitError(err, 1)
			} else {
				a.logger.Error("Failed to run command", zap.Error(err), zap.Strings("args", args))
				fmt.Fprintln(ctx.App.ErrWriter, err)
			}
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