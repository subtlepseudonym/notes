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
func EditNote(noteID int, body string, options EditOptions) error {
	note, err := files.GetNote(noteID)
	if err != nil {
		return errors.Wrap(err, "get note failed")
	}
	if !time.Unix(0, 0).Equal(note.Meta.Deleted) {
		note.Meta.Deleted = time.Unix(0, 0) // restore soft deleted notes
	}

	meta, err := files.GetMeta(Version)
	if err != nil {
		return errors.Wrap(err, "get meta failed")
	}

	if options.Title != "" {
		note.Meta.Title = options.Title
	}
	note.Body = body

	err = files.SaveNote(note)
	if err != nil {
		return errors.Wrap(err, "save note failed")
	}

	meta.Notes[note.Meta.ID] = note.Meta
	err = files.SaveMeta(meta)
	if err != nil {
		return errors.Wrap(err, "save meta failed")
	}

	return nil
}
