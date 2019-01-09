package notes

import (
	"time"

	"github.com/subtlepseudonym/notes/files"

	"github.com/pkg/errors"
)

type RemoveOptions struct {
	Hard bool // remove note file
}

// RemoveNote either soft or hard deletes a note
func RemoveNote(noteID int, options RemoveOptions) error {
	note, err := files.GetNote(noteID)
	if err != nil {
		return errors.Wrap(err, "get note failed")
	}

	if options.Hard {
		err = files.RemoveNote(noteID)
		if err != nil {
			return errors.Wrap(err, "remove note file failed")
		}
		return nil
	}

	note.Meta.Deleted = time.Now()
	err = files.SaveNote(note)
	if err != nil {
		return errors.Wrap(err, "save note failed")
	}

	meta, err := files.GetMeta(Version)
	if err != nil {
		return errors.Wrap(err, "get meta failed")
	}

	meta.Notes[note.Meta.ID] = note.Meta
	err = files.SaveMeta(meta)
	if err != nil {
		return errors.Wrap(err, "save meta failed")
	}

	return nil
}
