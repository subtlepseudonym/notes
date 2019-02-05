package notes

import (
	"time"

	"github.com/subtlepseudonym/notes/files"

	"github.com/pkg/errors"
)

// EditOptions defines the options available for modifying the behvaior
// of note editing
type EditOptions struct {
	Title string
}

// EditNote sets a new body and options for an existing note. It restores
// notes that were soft deleted.
func EditNote(note *files.Note, options EditOptions, dal files.DAL) (*files.Note, *files.Meta, error) {
	meta, err := dal.GetMeta()
	if err != nil {
		return nil, meta, errors.Wrap(err, "get meta failed")
	}

	if !time.Unix(0, 0).Equal(note.Meta.Deleted) {
		note.Meta.Deleted = time.Unix(0, 0) // restore soft deleted notes
	}

	if options.Title != "" {
		note.Meta.Title = options.Title
	}

	err = dal.SaveNote(note)
	if err != nil {
		return note, meta, errors.Wrap(err, "save note failed")
	}

	meta.Notes[note.Meta.ID] = note.Meta
	err = dal.SaveMeta(meta)
	if err != nil {
		return note, meta, errors.Wrap(err, "save meta failed")
	}

	return note, meta, nil
}
