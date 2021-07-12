package main

import (
	"fmt"
	"sort"

	"github.com/urfave/cli"
	"go.uber.org/zap"
)

func (a *App) buildNotebookCommand() cli.Command {
	return cli.Command{
		Name:        "notebook",
		Aliases:     []string{"nb"},
		Usage:       "print current notebook or access notebook subcommands",
		Description: "Print current notebook or access commands to create, modify, remove, and list available notebooks",
		Action:      a.notebookAction,
		Subcommands: []cli.Command{
			a.createNotebook(),
			a.listNotebooks(),
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
		Name:    "new",
		Aliases: []string{"n"},
		Usage:   "create a new notebook",
		Action:  a.createNotebookAction,
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

	a.logger.Info("notebook created", zap.String("notebook", name))
	return nil
}

func (a *App) listNotebooks() cli.Command {
	return cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "list existing notebooks",
		Action:  a.listNotebooksAction,
	}
}

func (a *App) listNotebooksAction(ctx *cli.Context) error {
	var notebooks []string
	for _, notebook := range a.data.GetAllNotebooks() {
		notebooks = append(notebooks, notebook)
	}
	sort.Strings(notebooks)

	for _, notebook := range notebooks {
		fmt.Fprintln(ctx.App.Writer, "  ", notebook)
	}

	return nil
}

func (a *App) setNotebook() cli.Command {
	return cli.Command{
		Name:    "use",
		Aliases: []string{"set"},
		Usage:   "set current notebook",
		Action:  a.setNotebookAction,
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

	meta, err := a.data.GetMeta()
	if err != nil {
		return fmt.Errorf("get meta: %v", err)
	}
	a.meta = meta

	return nil
}

func (a *App) renameNotebook() cli.Command {
	return cli.Command{
		Name:    "rename",
		Aliases: []string{"mv"},
		Usage:   "rename notebook",
		Action:  a.renameNotebookAction,
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

	a.logger.Info(
		"renamed notebook",
		zap.String("from", oldName),
		zap.String("to", newName),
	)

	return nil
}
