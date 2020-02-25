package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/subtlepseudonym/notes"
	dalpkg "github.com/subtlepseudonym/notes/dal"
	"github.com/subtlepseudonym/notes/dal/cache"

	"github.com/Masterminds/semver"
	"github.com/chzyer/readline"
	"github.com/kballard/go-shellquote"
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
	defaultCacheCapacity   = 16
)

type App struct {
	*cli.App

	homeDir   string
	setupOnce sync.Once

	logger *zap.Logger
	dal    dalpkg.DAL
	meta   *notes.Meta
	index  notes.Index

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
		cli.UintFlag{
			Name:  "cache-capacity",
			Usage: "Cache capacity",
			Value: defaultCacheCapacity,
		},
		cli.StringFlag{
			Name:  "cache",
			Usage: "Cache note state with `CACHE_TYPE`. Only useful with large note sets in interactive mode",
		},
	}

	app.Commands = []cli.Command{
		app.buildDebugCommand(),
		app.buildListCommand(),
		app.buildNewCommand(),
		app.buildRemoveCommand(),
		app.buildEditCommand(),
		app.buildInfoCommand(),
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

func (a *App) setup(ctx *cli.Context) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}
	a.homeDir = home

	logger, err := a.initLogging(ctx)
	if err != nil {
		return fmt.Errorf("init logging: %v", err)
	}
	a.logger = logger

	dal, err := dalpkg.NewLocal(defaultNotesDirectory, Version) // FIXME: option to use different dal
	if err != nil {
		return fmt.Errorf("initialize dal: %v", err)
	}

	if ctx.Int("cache-capacity") == 0 {
		return fmt.Errorf("cache capacity must be non-zero")
	}

	switch strings.ToLower(ctx.String("cache")) {
	case "lru", "least-recently-used":
		a.dal = cache.NewNoteCache(dal, cache.LRU, ctx.Int("capacity"))
	case "rr", "random-replacement":
		a.dal = cache.NewNoteCache(dal, cache.RR, ctx.Int("capacity"))
	default:
		a.dal = dal
	}

	index, err := dal.GetIndex()
	if err != nil {
		return fmt.Errorf("get index: %v", err)
	}
	a.index = index

	meta, err := dal.GetMeta()
	if err != nil {
		return fmt.Errorf("get meta: %v", err)
	}
	a.meta = meta

	return nil
}

func (a *App) before(ctx *cli.Context) error {
	var err error
	a.setupOnce.Do(func() {
		err = a.setup(ctx)
	})
	if err != nil {
		ctx.App.Writer = ioutil.Discard // prevent help text and double err printing
		return err
	}

	err = a.checkMetaVersion()
	if err != nil {
		a.logger.Error("check meta version", zap.Error(err))
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

func (a *App) checkMetaVersion() error {
	// TODO: this functionality is very likely to change when the update and
	// upgrade commands are implemented
	if a.meta == nil {
		return fmt.Errorf("meta is nil")
	}

	appVersion, err := semver.NewVersion(a.Version)
	if err != nil {
		return fmt.Errorf("parse app version: %w", err)
	}

	metaVersion, err := semver.NewVersion(a.meta.Version)
	if err != nil {
		return fmt.Errorf("parse meta version: %w", err)
	}

	// automatically update meta if it's not a new major version
	if appVersion.GreaterThan(metaVersion) && appVersion.Major() == metaVersion.Major() {
		a.logger.Info("updating meta version", zap.String("oldVersion", a.meta.Version))

		a.meta = a.meta.UpdateVersion(appVersion.String())
		size, err := a.meta.ApproxSize()
		if err != nil {
			return fmt.Errorf("approximate meta size: %v", err)
		}

		a.meta.Size = size
		err = a.dal.SaveMeta(a.meta)
		if err != nil {
			return fmt.Errorf("update meta version: save meta: %v", err)
		}
	}

	return nil
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
