package dal

import (
	"github.com/subtlepseudonym/notes"
)

type DAL interface {
	ReadNote(string) (*notes.Note, error)
	WriteNote(*notes.Note) error
	DeleteNote(string) error
}
