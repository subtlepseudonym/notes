package notes

import (
	"errors"
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
	// error behavior
	GetMetaError    error
	SaveMetaError   error
	GetNoteError    error
	SaveNoteError   error
	RemoveNoteError error
}

func (d FakeDAL) GetMeta() (*files.Meta, error) {
	if d.GetMetaError != nil {
		return nil, d.GetMetaError
	}

	return d.meta, nil
}

func (d FakeDAL) SaveMeta(meta *files.Meta) error {
	if d.SaveMetaError != nil {
		return d.SaveMetaError
	}

	d.meta = meta
	return nil
}

func (d FakeDAL) GetNote(id int) (*files.Note, error) {
	if d.GetNoteError != nil {
		return nil, d.GetNoteError
	}

	return d.notes[id], nil
}

func (d FakeDAL) SaveNote(note *files.Note) error {
	if d.SaveNoteError != nil {
		return d.SaveNoteError
	}

	d.notes[note.Meta.ID] = note
	return nil
}

func (d FakeDAL) RemoveNote(id int) error {
	if d.RemoveNoteError != nil {
		return d.RemoveNoteError
	}

	delete(d.notes, id)
	return nil
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
	ErrorExpected error // only ErrorExpected == nil is checked
}

func TestNewNote(t *testing.T) {
	fixedTime := time.Unix(100, 0).UTC()
	timePatch := monkey.Patch(time.Now, func() time.Time { return fixedTime })
	defer timePatch.Unpatch()

	// FIXME: this could probably be more readable
	tests := []NewNoteTest{
		NewNoteTest{
			Name: "dal.GetMeta() fails",
			DAL: FakeDAL{
				GetMetaError: errors.New("GetMeta"),
			},
			ErrorExpected: errors.New("non-nil"),
		},

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
			ExpectedMeta: &files.Meta{
				LatestID: 1,
				Notes: map[int]files.NoteMeta{
					1: files.NoteMeta{
						ID:      1,
						Title:   fixedTime.Local().Format(time.RFC1123),
						Created: fixedTime,
						Deleted: time.Unix(0, 0),
					},
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
			ExpectedMeta: &files.Meta{
				LatestID: 1,
				Notes: map[int]files.NoteMeta{
					1: files.NoteMeta{
						ID:      1,
						Title:   "TEST",
						Created: fixedTime,
						Deleted: time.Unix(0, 0),
					},
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
			ExpectedMeta: &files.Meta{
				LatestID: 1,
				Notes: map[int]files.NoteMeta{
					1: files.NoteMeta{
						ID:      1,
						Title:   fixedTime.Local().Format(time.UnixDate),
						Created: fixedTime,
						Deleted: time.Unix(0, 0),
					},
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
			ExpectedMeta: &files.Meta{
				LatestID: 1,
				Notes: map[int]files.NoteMeta{
					1: files.NoteMeta{
						ID:      1,
						Title:   fixedTime.UTC().Format(time.RFC1123),
						Created: fixedTime,
						Deleted: time.Unix(0, 0),
					},
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
			ExpectedMeta: &files.Meta{
				LatestID: 1,
				Notes: map[int]files.NoteMeta{
					1: files.NoteMeta{
						ID:      1,
						Title:   fixedTime.Local().Format(time.RFC1123),
						Created: fixedTime,
						Deleted: time.Unix(0, 0),
					},
				},
			},
		},

		NewNoteTest{
			Name: "dal.SaveNote() fails",
			DAL: FakeDAL{
				meta: &files.Meta{
					LatestID: 10, // something non-zero
				},
				SaveNoteError: errors.New("SaveNote"),
			},
			ExpectedMeta: &files.Meta{
				LatestID: 10,
			},
			ErrorExpected: errors.New("non-nil"),
		},

		NewNoteTest{
			Name: "non-unique note ID", // could happen on corrupt meta object
			DAL: FakeDAL{
				meta: &files.Meta{
					LatestID: 1,
					Notes: map[int]files.NoteMeta{
						2: files.NoteMeta{},
					},
				},
				notes: make(map[int]*files.Note),
			},
			ExpectedMeta: &files.Meta{
				LatestID: 1,
				Notes: map[int]files.NoteMeta{
					2: files.NoteMeta{},
				},
			},
			ErrorExpected: errors.New("non-nil"),
		},

		NewNoteTest{
			Name: "dal.SaveMeta() fails",
			DAL: FakeDAL{
				meta: &files.Meta{
					Notes: make(map[int]files.NoteMeta),
				},
				notes:         make(map[int]*files.Note),
				SaveMetaError: errors.New("SaveMeta"),
			},
			ExpectedMeta: &files.Meta{
				LatestID: 1,
				Notes: map[int]files.NoteMeta{
					1: files.NoteMeta{
						ID:      1,
						Title:   fixedTime.Local().Format(time.RFC1123),
						Created: fixedTime,
						Deleted: time.Unix(0, 0),
					},
				},
			},
			ErrorExpected: errors.New("non-nil"),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			note, meta, err := NewNote(test.Body, test.Options, test.DAL)
			if test.ErrorExpected == nil && err != nil {
				t.Error(err)
			}

			if diff := deep.Equal(note, test.ExpectedNote); diff != nil {
				t.Error(diff)
			}
			if diff := deep.Equal(meta, test.ExpectedMeta); diff != nil {
				t.Error(diff)
			}

			if err != nil {
				return
			}

			savedNote, err := test.DAL.GetNote(note.Meta.ID)
			if err != nil {
				t.Errorf("test.DAL.GetNote failed: %s", err)
				t.FailNow()
			}

			if diff := deep.Equal(savedNote, test.ExpectedNote); diff != nil {
				t.Error(diff)
			}

			savedMeta, err := test.DAL.GetMeta()
			if err != nil {
				t.Errorf("test.DAL.GetMeta failed: %s", err)
				t.FailNow()
			}

			if diff := deep.Equal(savedMeta, test.ExpectedMeta); diff != nil {
				t.Error(diff)
			}
		})
	}
}
