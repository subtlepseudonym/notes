package main

import (
	"fmt"

	"github.com/urfave/cli"
)

func (a *App) buildNotebookCommand() cli.Command {
	return cli.Command{
		Name:        "notebook",
		ShortName:   "nb",
		Usage:       "create or modify notebooks",
		Description: "Create, modify, remove, and list available notebooks",
		Action:      a.notebookAction,
		Subcommands: []cli.Command{
			a.createNotebook(),
			a.setNotebook(),
			a.renameNotebook(),
		},
	}
}

func (a *App) notebookAction(ctx *cli.Context) error {
	fmt.Fprintln(ctx.App.Writer, a.data.GetNotebook())
	return nil
}

func (a *App) createNotebook() cli.Command {
	return cli.Command{
		Name:      "new",
		ShortName: "n",
		Usage:     "create a new notebook",
		Action:    a.createNotebookAction,
	}
}

func (a *App) createNotebookAction(ctx *cli.Context) error {
	if !ctx.Args().Present() {
		return fmt.Errorf("usage: notebook name required")
	}
	name := ctx.Args().First()

	err := a.data.CreateNotebook(name)
	if err != nil {
		return fmt.Errorf("create notebook: %v", err)
	}

	err = a.data.SetNotebook(name)
	if err != nil {
		return fmt.Errorf("set notebook: %v", err)
	}

	return nil
}

func (a *App) setNotebook() cli.Command {
	return cli.Command{
		Name:   "set",
		Usage:  "set current notebook",
		Action: a.setNotebookAction,
	}
}

func (a *App) setNotebookAction(ctx *cli.Context) error {
	if !ctx.Args().Present() {
		return fmt.Errorf("usage: notebook name required")
	}
	name := ctx.Args().First()

	err := a.data.SetNotebook(name)
	if err != nil {
		return fmt.Errorf("set notebook: %v", err)
	}

	return nil
}

func (a *App) renameNotebook() cli.Command {
	return cli.Command{
		Name:      "rename",
		ShortName: "mv",
		Usage:     "rename notebook",
		Action:    a.renameNotebookAction,
	}
}

func (a *App) renameNotebookAction(ctx *cli.Context) error {
	if !ctx.Args().Present() {
		return fmt.Errorf("usage: notebook name required")
	}

	oldName := ctx.Args().Get(0)
	newName := ctx.Args().Get(1)

	err := a.data.RenameNotebook(oldName, newName)
	if err != nil {
		return fmt.Errorf("rename notebook: %v", err)
	}

	err = a.data.SetNotebook(newName)
	if err != nil {
		return fmt.Errorf("set notebook: %v", err)
	}

	return nil
}
