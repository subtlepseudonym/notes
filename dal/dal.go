package dal

import (
	"github.com/subtlepseudonym/notes"
)

type DAL interface {
	CreateNote(*notes.Note) error
	ReadNote(string) (*notes.Note, error)
	UpdateNote(string, *notes.Note) (*notes.Note, error)
	DeleteNote(string) error
}
