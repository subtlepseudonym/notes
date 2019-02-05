package notes

import (
	"fmt"
	"testing"
	"time"

	"github.com/subtlepseudonym/notes/files"

	"github.com/bouk/monkey"
	"github.com/go-test/deep"
)

// TODO: move this into a more general location
// TODO: it's going to get used it edit_test.go etc
type FakeDAL struct {
	meta  *files.Meta
	notes map[int]*files.Note
}

func (d FakeDAL) GetMeta() (*files.Meta, error) {
	return d.meta, nil
}

func (d FakeDAL) SaveMeta(meta *files.Meta) error {
	d.meta = meta
	return nil
}

func (d FakeDAL) GetNote(id int) (*files.Note, error) {
	note, exists := d.notes[id]
	if !exists {
		return nil, fmt.Errorf("no note with id \"%d\"", id)
	}

	return note, nil
}

func (d FakeDAL) SaveNote(note *files.Note) error {
	d.notes[note.Meta.ID] = note
	return nil
}

func (d FakeDAL) RemoveNote(id int) error {
	delete(d.notes, id)
	return nil
}

// NewNoteTest defines the input arguments and expected output
// of each NewNote subtest
type NewNoteTest struct {
	Name    string
	Body    string
	Options NoteOptions
	DAL     files.DAL

	ExpectedNote *files.Note
}

func TestNewNote(t *testing.T) {
	fixedTime := time.Unix(100, 0).UTC()
	timePatch := monkey.Patch(time.Now, func() time.Time { return fixedTime })
	defer timePatch.Unpatch()

	tests := []NewNoteTest{
		NewNoteTest{
			Name: "with existing note",
			DAL: FakeDAL{
				meta: &files.Meta{
					LatestID: 6,
					Notes:    make(map[int]files.NoteMeta),
				},
				notes: make(map[int]*files.Note),
			},
			ExpectedNote: &files.Note{
				Meta: files.NoteMeta{
					ID:      7,
					Title:   fixedTime.Local().Format(time.RFC1123),
					Created: fixedTime,
					Deleted: time.Unix(0, 0),
				},
			},
		},
		NewNoteTest{
			Name: "default title",
			DAL: FakeDAL{
				meta: &files.Meta{
					Notes: make(map[int]files.NoteMeta),
				},
				notes: make(map[int]*files.Note),
			},
			ExpectedNote: &files.Note{
				Meta: files.NoteMeta{
					ID:      1,
					Title:   fixedTime.Local().Format(time.RFC1123),
					Created: fixedTime,
					Deleted: time.Unix(0, 0),
				},
			},
		},
		NewNoteTest{
			Name: "custom title",
			Options: NoteOptions{
				Title: "TEST",
			},
			DAL: FakeDAL{
				meta: &files.Meta{
					Notes: make(map[int]files.NoteMeta),
				},
				notes: make(map[int]*files.Note),
			},
			ExpectedNote: &files.Note{
				Meta: files.NoteMeta{
					ID:      1,
					Title:   "TEST",
					Created: fixedTime,
					Deleted: time.Unix(0, 0),
				},
			},
		},
		NewNoteTest{
			Name: "custom date title format",
			Options: NoteOptions{
				DateTitleFormat: time.UnixDate,
			},
			DAL: FakeDAL{
				meta: &files.Meta{
					Notes: make(map[int]files.NoteMeta),
				},
				notes: make(map[int]*files.Note),
			},
			ExpectedNote: &files.Note{
				Meta: files.NoteMeta{
					ID:      1,
					Title:   fixedTime.Local().Format(time.UnixDate),
					Created: fixedTime,
					Deleted: time.Unix(0, 0),
				},
			},
		},
		NewNoteTest{
			Name: "custom date title location",
			Options: NoteOptions{
				DateTitleLocation: "UTC",
			},
			DAL: FakeDAL{
				meta: &files.Meta{
					Notes: make(map[int]files.NoteMeta),
				},
				notes: make(map[int]*files.Note),
			},
			ExpectedNote: &files.Note{
				Meta: files.NoteMeta{
					ID:      1,
					Title:   fixedTime.UTC().Format(time.RFC1123),
					Created: fixedTime,
					Deleted: time.Unix(0, 0),
				},
			},
		},
		NewNoteTest{
			Name: "with body",
			Body: "very important note!",
			DAL: FakeDAL{
				meta: &files.Meta{
					Notes: make(map[int]files.NoteMeta),
				},
				notes: make(map[int]*files.Note),
			},
			ExpectedNote: &files.Note{
				Meta: files.NoteMeta{
					ID:      1,
					Title:   fixedTime.Local().Format(time.RFC1123),
					Created: fixedTime,
					Deleted: time.Unix(0, 0),
				},
				Body: "very important note!",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			note, _, err := NewNote(test.Body, test.Options, test.DAL)
			if err != nil {
				t.Errorf("NewNote failed: %s", err)
				t.FailNow()
			}

			if diff := deep.Equal(note, test.ExpectedNote); diff != nil {
				t.Error(diff)
			}

			savedNote, err := test.DAL.GetNote(note.Meta.ID)
			if err != nil {
				t.Errorf("test.DAL.GetNote failed: %s", err)
				t.FailNow()
			}

			if diff := deep.Equal(savedNote, test.ExpectedNote); diff != nil {
				t.Error(diff)
			}
		})
	}
}
