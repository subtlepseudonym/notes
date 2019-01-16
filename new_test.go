package notes

import (
	"testing"
	"time"
)

func TestNewNote(t *testing.T) {
	options := map[string]NoteOptions{
		"default title": NoteOptions{},
		"custom title": NoteOptions{
			Title: "TEST",
		},
		"custom date title format": NoteOptions{
			DateTitleFormat: time.Kitchen,
		},
		"custom date title location": NoteOptions{
			DateTitleLocation: time.UTC,
		},
	}

	// TODO: subtests
	// TODO: above options are good for testing title generation
	// TODO: may want to break these out into body tests and meta tests
}
