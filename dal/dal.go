package dal

import (
	"github.com/subtlepseudonym/notes"
)

// DAL interfaces the method by which we access the source of
// Meta and Note objects
type DAL interface {
	GetMeta() (*notes.Meta, error)
	SaveMeta(*notes.Meta) error

	CreateNotebook(string) error
	GetNotebook() string
	SetNotebook(string) error
	RenameNotebook(string, string) error
	RemoveNotebook(string, bool) error

	GetNoteMeta(int) (*notes.NoteMeta, error)
	GetAllNoteMetas() (map[int]notes.NoteMeta, error)

	GetNote(int) (*notes.Note, error)
	SaveNote(*notes.Note) error
	RemoveNote(int) error
}
