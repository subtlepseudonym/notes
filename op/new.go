package operations

import (
	"fmt"
	"time"

	"github.com/subtlepseudonym/notes"
	"github.com/subtlepseudonym/notes/dal"
)

const (
	defaultDateTitleFormat = time.RFC1123
)

// NewNoteOptions provides values by which to alter the Note created by NewNote
type NewNoteOptions struct {
	Title        string `json:"title"`
	DateFormat   string `json:"dateFormat"`
	DateLocation string `json:"dateLocation"`
}

// NewNote creates a new note object and populates it according to the provided options
// and updates the state accordingly
func NewNote(data dal.DAL, options NewNoteOptions) error {
	index, err := data.GetIndex()
	if err != nil {
		return fmt.Errorf("get index: %v", err)
	}

	meta, err := data.GetMeta()
	if err != nil {
		return fmt.Errorf("get meta: %v", err)
	}

	// FIXME: use UUIDs
	newNoteID := meta.LatestID + 1
	if _, ok := index[newNoteID]; ok {
		return fmt.Errorf("note ID: must be unique")
	}

	title := options.Title
	if title == "" {
		format := defaultDateTitleFormat
		if options.DateFormat != "" {
			format = options.DateFormat
		}

		loc := time.UTC
		if options.DateLocation != "" {
			l, err := time.LoadLocation(options.DateLocation)
			if err == nil {
				loc = l
			}
		}

		title = time.Now().In(loc).Format(format)
	}

	note := &notes.Note{
		Meta: notes.NoteMeta{
			ID:      newNoteID,
			Title:   title,
			Created: notes.JSONTime{time.Now()},
			Deleted: notes.JSONTime{time.Unix(0, 0)},
		},
	}

	err = data.SaveNote(note)
	if err != nil {
		return fmt.Errorf("save note: %v", err)
	}

	index[note.Meta.ID] = note.Meta
	err = data.SaveIndex(index)
	if err != nil {
		return fmt.Errorf("save index: %v", err)
	}

	meta.LatestID = note.Meta.ID
	metaSize, err := meta.ApproxSize()
	if err != nil {
		return fmt.Errorf("approximate meta size: %w", err)
	}
	meta.Size = metaSize

	err = data.SaveMeta(meta)
	if err != nil {
		return fmt.Errorf("save meta: %v", err)
	}

	return nil
}
