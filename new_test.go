package notes

import (
	"testing"
	"time"

	"github.com/bouk/monkey"
	"github.com/go-test/deep"
	"github.com/subtlepseudonym/notes/files"
)

// NewNoteTest defines the input arguments and expected output
// of each NewNote subtest
type NewNoteTest struct {
	Name    string
	Body    string
	Options NoteOptions
	Meta    *files.Meta

	ExpectedNote files.Note
}

func TestNewNote(t *testing.T) {
	fixedTime := time.Unix(100, 0).UTC()
	timePatch := monkey.Patch(time.Now, func() time.Time { return fixedTime })
	defer timePatch.Unpatch()

	tests := []NewNoteTest{
		NewNoteTest{
			Name:    "default title, empty body",
			Options: NoteOptions{},
			Meta: &files.Meta{
				Version:  "v0.0.0",
				LatestID: 0,
				Notes:    make(map[int]files.NoteMeta),
			},
			ExpectedNote: files.Note{
				Meta: files.NoteMeta{
					ID:      1,
					Title:   fixedTime.Local().Format(time.RFC1123),
					Created: fixedTime,
					Deleted: time.Unix(0, 0),
				},
			},
		},
		NewNoteTest{
			Name: "custom title, empty body",
			Options: NoteOptions{
				Title: "TEST",
			},
			Meta: &files.Meta{
				Version:  "v0.0.0",
				LatestID: 0,
				Notes:    make(map[int]files.NoteMeta),
			},
			ExpectedNote: files.Note{
				Meta: files.NoteMeta{
					ID:      1,
					Title:   "TEST",
					Created: fixedTime,
					Deleted: time.Unix(0, 0),
				},
			},
		},
		NewNoteTest{
			Name: "custom date title format, empty body",
			Options: NoteOptions{
				DateTitleFormat: time.UnixDate,
			},
			Meta: &files.Meta{
				Version:  "v0.0.0",
				LatestID: 0,
				Notes:    make(map[int]files.NoteMeta),
			},
			ExpectedNote: files.Note{
				Meta: files.NoteMeta{
					ID:      1,
					Title:   fixedTime.Local().Format(time.UnixDate),
					Created: fixedTime,
					Deleted: time.Unix(0, 0),
				},
			},
		},
		NewNoteTest{
			Name: "custom date title location, empty body",
			Options: NoteOptions{
				DateTitleLocation: "UTC",
			},
			Meta: &files.Meta{
				Version:  "v0.0.0",
				LatestID: 0,
				Notes:    make(map[int]files.NoteMeta),
			},
			ExpectedNote: files.Note{
				Meta: files.NoteMeta{
					ID:      1,
					Title:   fixedTime.UTC().Format(time.RFC1123),
					Created: fixedTime,
					Deleted: time.Unix(0, 0),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			note, _, err := NewNote(test.Body, test.Options, test.Meta)
			if err != nil {
				t.Errorf("NewNote failed: %s\n", err)
				t.FailNow()
			}

			if diff := deep.Equal(note, test.ExpectedNote); diff != nil {
				t.Error(diff)
			}
		})
	}
	// TODO: subtests
	// TODO: above options are good for testing title generation
	// TODO: may want to break these out into body tests and meta tests
}
