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
// of new note creation
type NoteOptions struct {
	Title string

	DateTitleFormat   string
	DateTitleLocation string
}

// NewNote generates a new Note and populates the metadata, saves it to
// file, and returns it
func NewNote(body string, options NoteOptions, dal files.DAL) (*files.Note, *files.Meta, error) {
	meta, err := dal.GetMeta()
	if err != nil {
		return nil, nil, errors.Wrap(err, "get meta failed")
	}
	newNoteID := meta.LatestID + 1

	var title string
	if options.Title != "" {
		title = options.Title
	} else {
		title = generateDateTitle(options.DateTitleFormat, options.DateTitleLocation)
	}

	note := &files.Note{
		Meta: files.NoteMeta{
			ID:      newNoteID,
			Title:   title,
			Created: time.Now().UTC(),
			Deleted: time.Unix(0, 0),
		},
		Body: body,
	}

	err = dal.SaveNote(note)
	if err != nil {
		// FIXME: persist the note somewhere if saving it fails
		return nil, meta, errors.Wrap(err, "save note failed")
	}

	_, exists := meta.Notes[note.Meta.ID]
	if exists {
		return nil, meta, errors.New("note ID is not unique")
	}

	meta.Notes[note.Meta.ID] = note.Meta
	meta.LatestID = note.Meta.ID
	err = dal.SaveMeta(meta)
	if err != nil {
		return nil, meta, errors.Wrap(err, "save meta failed")
	}

	return note, meta, nil
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
