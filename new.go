package notes

import (
	"time"

	"github.com/subtlepseudonym/notes/files"

	"github.com/pkg/errors"
)

const (
	defaultDateTitleFormat   = time.RFC1123
	defaultDateTitleLocation = "Local"
)

// NoteOptions defines the set of options for modifying the behavior
// new note creation
type NoteOptions struct {
	Title string

	DateTitleFormat   string
	DateTitleLocation string
}

// NewNote generates a new Note and populates the metadata, saves it to
// file, and returns it
func NewNote(options NoteOptions) (*files.Note, error) {
	meta, err := files.GetMeta(Version)
	if err != nil {
		return nil, errors.Wrap(err, "get meta failed")
	}

	newNoteID := 1
	if len(meta.Notes) > 0 {
		newNoteID = meta.LatestID + 1
	}

	var title string
	if options.Title != "" {
		title = options.Title
	} else {
		title = generateDateTitle(options.DateTitleFormat, options.DateTitleLocation)
	}

	note := files.Note{
		Meta: files.NoteMeta{
			ID:      newNoteID,
			Title:   title,
			Created: time.Now().UTC(),
		},
	}

	err = files.SaveNote(note)
	if err != nil {
		// FIXME: persist the note somewhere if saving it fails
		return nil, errors.Wrap(err, "save note failed")
	}

	_, exists := meta.Notes[note.Meta.ID]
	if exists {
		return nil, errors.New("note ID is not unique")
	}

	meta.Notes[note.Meta.ID] = note.Meta
	meta.LatestID = note.Meta.ID
	err = files.SaveMeta(meta)
	if err != nil {
		return nil, errors.Wrap(err, "save meta failed")
	}

	return &note, nil
}

func generateDateTitle(userFormat, userLocation string) string {
	locStr := defaultDateTitleLocation
	if userLocation != "" {
		locStr = userLocation
	}
	var loc *time.Location
	l, err := time.LoadLocation(locStr)
	if err != nil {
		// TODO: log error
		loc = time.UTC
	} else {
		loc = l
	}

	format := defaultDateTitleFormat
	if userFormat != "" {
		format = userFormat
	}

	return time.Now().In(loc).Format(format)
}
