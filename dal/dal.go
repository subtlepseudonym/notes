package dal

import (
	"github.com/subtlepseudonym/notes"
)

// DAL interfaces the method by which we access the source of
// Meta and Note objects
type DAL interface {
	GetIndex() (notes.Index, error)
	SaveIndex(notes.Index) error

	GetMeta() (*notes.Meta, error)
	SaveMeta(*notes.Meta) error

	GetNote(int) (*notes.Note, error)
	SaveNote(*notes.Note) error
	RemoveNote(int) error
}
