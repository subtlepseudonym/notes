package main

import (
	"fmt"
	"os"

	"github.com/chzyer/readline"
	"github.com/urfave/cli"
)

// Set at compile time
var (
	Version  = "v0.0.0"
	Revision = "git_revision"
)

const (
	defaultEditor   = "vim"
	historyFilePath = ".nts_history"
)

func main() {
	os.Setenv("_CLI_ZSH_AUTOCOMPLETE_HACK", "1")

	app := cli.NewApp()

	app.Usage = "write and organize notes"
	app.Description = "notes is intended to make it easy to jot down stream of consciousness notes while working in the command line and automatically back those notes up to a remote server"

	app.Version = Version
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Connor Demille",
			Email: "subtlepseudonym@gmail.com",
		},
	}

	app.EnableBashCompletion = true
	app.ErrWriter = os.Stderr

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

func mainAction(ctx *cli.Context) {
	dal, err := notes.NewDefaultDAL(Version) // FIXME: option to use different dal
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "initialize dal failed"), 1)
	}

	config := *readline.Config{
		Prompt:       ctx.App.Name + "> ",
		HistoryFile:  dal.notesDirectoryPath + "/" + historyFilePath,
		HistorySearchFold: true,
		Autocomplete: readline.NewPrefixCompleter(buildPrefixCompleter(ctx.App.Commands)),
		InterruptPrompt: "^C",
		EOFPrompt: "exit",
	}

	reader, err := readline.NewEx(config)
	if err != nil {
		return cli.NewExitError(errors.Wrap(err, "create reader failed"), 1)
	}
}

func buildPrefixCompleter(cmds []cli.Command) []readline.PrefixCompleter {
	var completers []readline.PrefixCompleter
	for _, cmd := range cmds {
		completers = append(completers, readline.PcItem(cmd.Name, buildPrefixCompleter([]cli.Command(cmd.Subcommands))))
	}
	return completers
}
