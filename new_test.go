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

type ErrorDALGetMeta struct {
	FakeDAL
}

func (d ErrorDALGetMeta) GetMeta() (*files.Meta, error) {
	return nil, fmt.Errorf("no meta for you")
}

type ErrorDALGetNote struct {
	FakeDAL
}

func (d ErrorDALGetNote) GetNote(id int) (*files.Note, error) {
	return nil, fmt.Errorf("no note for you")
}

// NewNoteTest defines the input arguments and expected output
// of each NewNote subtest
type NewNoteTest struct {
	Name string
	// args
	Body    string
	Options NoteOptions
	DAL     files.DAL
	// output
	ExpectedNote  *files.Note
	ExpectedMeta  *files.Meta
	ExpectedError error
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
			ExpectedMeta: &files.Meta{
				LatestID: 7,
				Notes: map[int]files.NoteMeta{
					7: files.NoteMeta{
						ID:      7,
						Title:   fixedTime.Local().Format(time.RFC1123),
						Created: fixedTime,
						Deleted: time.Unix(0, 0),
					},
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
			note, meta, err := NewNote(test.Body, test.Options, test.DAL)

			if diff := deep.Equal(err, test.ExpectedError); diff != nil {
				t.Error(diff)
			}
			if diff := deep.Equal(note, test.ExpectedNote); diff != nil {
				t.Error(diff)
			}
			if diff := deep.Equal(meta, test.ExpectedMeta); diff != nil {
				t.Error(diff)
			}

			if _, ok := test.DAL.(ErrorDALGetNote); !ok {
				savedNote, err := test.DAL.GetNote(note.Meta.ID)
				if err != nil {
					t.Errorf("test.DAL.GetNote failed: %s", err)
					t.FailNow()
				}

				if diff := deep.Equal(savedNote, test.ExpectedNote); diff != nil {
					t.Error(diff)
				}
			}
		})
	}
}
