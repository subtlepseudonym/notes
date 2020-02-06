// +build debug

package main

import (
	"encoding/json"
	"fmt"
	"path"
	"strconv"

	"github.com/urfave/cli"
)

func (a *App) buildDebugCommand() cli.Command {
	return cli.Command{
		Name:        "debug",
		Usage:       "access debugging tools",
		Description: "Access lower-level structures, implementation details, and other debugging utilities. The behavior of this command and its subcommands are subject to breaking changes across non-major releases",
		Subcommands: []cli.Command{
			debugGetNote(),
			debugGetMeta(),
			debugGetIndex(),
			debugRebuildIndex(),
		},
	}
}

func debugGetNote() cli.Command {
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

	note, err := a.dal.GetNote(noteID)
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

func debugGetMeta() cli.Command {
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
		meta, err = a.dal.GetMeta()
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

func debugGetIndex() cli.Command {
	return cli.Command{
		Name:        "get-index",
		Usage:       "print index structure",
		Description: "Print the contents of the index as a json object",
		Action:      a.getIndexAction,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "in-memory",
				Usage: "get index that's currently in memory",
			},
		},
	}
}

func (a *App) getIndexAction(ctx *cli.Context) error {
	index := a.index
	if !ctx.Bool("in-memory") {
		var err error
		index, err = a.dal.GetIndex()
		if err != nil {
			return fmt.Errorf("get index from dal: %w", err)
		}
	}

	b, err := json.Marshal(index)
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}

	_, err = ctx.AppWriter.Write(b)
	if err != nil {
		return fmt.Errorf("write to app writer: %w", err)
	}
	return nil
}

func debugRebuildIndex() cli.Command {
	return cli.Command{
		Name:        "rebuild-index",
		Usage:       "rebuild index file",
		Description: "Create the index file from scratch. This is only useful when use the local DAL",
		Action:      a.rebuildIndexAction(),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "force",
				Usage: "rebuild index even if there is no local DAL present in the home directory",
			},
			cli.BoolFlag{
				Name:  "no-backup",
				Usage: "don't create a backup of the index file before writing to disk",
			},
			cli.IntFlag{
				Name:  "capacity",
				Usage: "index capacity",
				Value: 0,
			},
		},
	}
}

func (a *App) rebuildIndexAction(ctx *cli.Context) error {
	notesDir := path.Join(a.homedir, defaultNotesDirectory)
	a.logger = a.logger.Named("rebuild-index")

	if _, err := os.Stat(notesDir); os.IsNotExist(err) {
		if !ctx.Bool("force") {
			return fmt.Errorf("local DAL not found: %w\nuse --force if you'd like to rebuild anyway", err)
		}

		err = os.Mkdir(notesDir, osModDir|os.FileMode(0700))
		if err != nil {
			return fmt.Errorf("create notes directory: %w", err)
		}
	}

	indexCapacity := ctx.Int("capacity")
	if indexCapacity < 0 {
		return fmt.Errorf("index capacity cannot be less than zero")
	}

	index := notes.NewIndex(indexCapacity)
	infos, err := ioutil.ReadDir(notesDir)
	if err != nil {
		return fmt.Errorf("read notes directory: %w", err)
	}

	nameRegex, err := regexp.Compile(`[0-9]{6}`)
	if err != nil {
		return fmt.Errorf("compile regex: %w", err)
	}

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
			a.logger.Error(fmt.Sprintf("decode note: %w", err))
			noteFile.Close()
			continue
		}

		index[note.Meta.ID] = note.Meta
		noteFile.Close()
	}

	// FIXME: this will need to be updated if the default index filename is ever changed or
	// 		  if the option to rename the file is ever provided
	indexPath := path.Join(a.homedir, "index")
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

	if backupIndex {
		err = os.Remove(indexPath + ".rebuild.bak")
		if err != nil {
			return fmt.Errorf("remove index backup: %w", err)
		}
	}

	return nil
}
