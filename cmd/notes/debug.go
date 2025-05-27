//go:build debug

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"

	"github.com/subtlepseudonym/notes"
	"github.com/subtlepseudonym/notes/dal"
	"github.com/urfave/cli"
)

func (a *App) buildDebugCommand() cli.Command {
	return cli.Command{
		Name:        "debug",
		Usage:       "access debugging tools",
		Description: "Access lower-level structures, implementation details, and other debugging utilities. The behavior of this command and its subcommands are subject to breaking changes across non-major releases",
		Subcommands: []cli.Command{
			a.getNote(),
			a.getMeta(),
			a.getNoteMetas(),
			a.rebuildIndex(),
		},
	}
}

func (a *App) getNote() cli.Command {
	return cli.Command{
		Name:        "get-note",
		Usage:       "print note structure",
		Description: "Print the contents of a note as a json object",
		ArgsUsage:   "<noteID>",
		Action:      a.getNoteAction,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "no-body",
				Usage: "don't include note body",
			},
		},
	}
}

func (a *App) getNoteAction(ctx *cli.Context) error {
	if !ctx.Args().Present() {
		return fmt.Errorf("usage: noteID argument required")
	}
	n, err := strconv.ParseInt(ctx.Args().First(), 16, 64)
	if err != nil {
		return fmt.Errorf("parse noteID argument: %w", err)
	}
	noteID := int(n)

	note, err := a.data.GetNote(noteID)
	if err != nil {
		return fmt.Errorf("get note: %w", err)
	}

	if ctx.Bool("no-body") {
		note.Body = "[EXCLUDED]"
	}

	b, err := json.Marshal(note)
	if err != nil {
		return fmt.Errorf("marshal note: %w", err)
	}

	_, err = ctx.App.Writer.Write(b)
	if err != nil {
		return fmt.Errorf("write to app writer: %w", err)
	}
	return nil
}

func (a *App) getMeta() cli.Command {
	return cli.Command{
		Name:        "get-meta",
		Usage:       "print meta structure",
		Description: "Print the contents of the meta as a json object",
		Action:      a.getMetaAction,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "in-memory",
				Usage: "get meta that's currently in memory",
			},
		},
	}
}

func (a *App) getMetaAction(ctx *cli.Context) error {
	meta := a.meta
	if !ctx.Bool("in-memory") {
		var err error
		meta, err = a.data.GetMeta()
		if err != nil {
			return fmt.Errorf("get meta from dal: %w", err)
		}
	}

	b, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal meta: %w", err)
	}

	_, err = ctx.App.Writer.Write(b)
	if err != nil {
		return fmt.Errorf("write to app writer: %w", err)
	}
	return nil
}

func (a *App) getNoteMetas() cli.Command {
	return cli.Command{
		Name:        "get-note-metas",
		Usage:       "print note meta objects for",
		Description: "Print a map of note IDs to their meta information",
		Action:      a.getNoteMetasAction,
	}
}

func (a *App) getNoteMetasAction(ctx *cli.Context) error {
	index, err := a.data.GetAllNoteMetas()
	if err != nil {
		return fmt.Errorf("get note metas: %w", err)
	}
	b, err := json.Marshal(index)
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}

	_, err = ctx.App.Writer.Write(b)
	if err != nil {
		return fmt.Errorf("write to app writer: %w", err)
	}
	return nil
}

func (a *App) rebuildIndex() cli.Command {
	return cli.Command{
		Name:        "rebuild-index",
		Usage:       "rebuild index file",
		Description: "Rebuild the local DAL index file from scratch",
		Action:      a.rebuildIndexAction,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "no-backup",
				Usage: "don't create a backup of the index file before writing to disk",
			},
		},
	}
}

func (a *App) rebuildIndexAction(ctx *cli.Context) error {
	notebook := a.data.GetNotebook()
	a.logger = a.logger.Named(notebook).Named("rebuild-index")

	notesDir := path.Join(a.homeDir, defaultNotesDirectory)
	if _, err := os.Stat(notesDir); os.IsNotExist(err) {
		return fmt.Errorf("local DAL not found: %w\nuse --force if you'd like to rebuild anyway", err)
	}

	infos, err := ioutil.ReadDir(notesDir)
	if err != nil {
		return fmt.Errorf("read notes directory: %w", err)
	}

	nameRegex, err := regexp.Compile(`[0-9]{6}`)
	if err != nil {
		return fmt.Errorf("compile regex: %w", err)
	}

	index := make(map[int]notes.NoteMeta)
	for _, info := range infos {
		if info.IsDir() || !nameRegex.MatchString(info.Name()) {
			continue
		}

		noteFilename := path.Join(notesDir, info.Name())
		noteFile, err := os.Open(noteFilename)
		if err != nil {
			a.logger.Error(fmt.Sprintf("open %s: %v", noteFilename, err))
			continue
		}

		var note notes.Note
		err = json.NewDecoder(noteFile).Decode(&note)
		if err != nil {
			a.logger.Error(fmt.Sprintf("decode note: %s", err))
			noteFile.Close()
			continue
		}

		index[note.Meta.ID] = note.Meta
		noteFile.Close()
	}

	// FIXME: this will need to be updated if the default index filename is ever changed or
	// 		  if the option to rename the file is ever provided
	indexPath := path.Join(a.homeDir, defaultNotesDirectory, "index")
	backupIndex := !ctx.Bool("no-backup")
	if backupIndex {
		err := os.Rename(indexPath, indexPath+".rebuild.bak")
		if err != nil {
			return fmt.Errorf("backup index: %w\n use --no-backup if you'd like to build anyway", err)
		}
	}

	indexFile, err := os.Create(indexPath)
	if err != nil {
		if backupIndex {
			err = os.Rename(indexPath+".rebuild.bak", indexPath)
			if err != nil {
				return fmt.Errorf("restore index backup: %w", err)
			}
		}
		return fmt.Errorf("create new index file: %w", err)
	}

	err = json.NewEncoder(indexFile).Encode(index)
	if err != nil {
		return fmt.Errorf("encode index file: %w", err)
	}

	err = indexFile.Close()
	if err != nil {
		return fmt.Errorf("close index file: %w", err)
	}

	local, err := dal.NewLocal(defaultNotesDirectory, a.meta.Version)
	if err != nil {
		return fmt.Errorf("new local dal: %w", err)
	}
	a.data = local

	return nil
}
