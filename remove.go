package notes

import (
	"time"

	"github.com/subtlepseudonym/notes/files"

	"github.com/pkg/errors"
)

// RemoveOptions defines all the options available for editiong the behavior of
// note removal
type RemoveOptions struct {
	Hard bool // remove note file
}

// RemoveNote either soft or hard deletes a note
func RemoveNote(noteID int, options RemoveOptions, dal files.DAL) error {
	note, err := dal.GetNote(noteID)
	if err != nil {
		return errors.Wrap(err, "get note failed")
	}

	meta, err := dal.GetMeta()
	if err != nil {
		return errors.Wrap(err, "get meta failed")
	}

	if options.Hard {
		err = dal.RemoveNote(noteID)
		if err != nil {
			return errors.Wrap(err, "remove note file failed")
		}

		delete(meta.Notes, note.Meta.ID)
	} else {
		note.Meta.Deleted = time.Now()
		err = dal.SaveNote(note)
		if err != nil {
			return errors.Wrap(err, "save note failed")
		}

		meta.Notes[note.Meta.ID] = note.Meta
	}

	err = dal.SaveMeta(meta)
	if err != nil {
		return errors.Wrap(err, "save meta failed")
	}

	return nil
}
