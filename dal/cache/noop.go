package cache

import (
	"errors"

	"github.com/subtlepseudonym/notes/dal"
)

type noop struct {
	dal.DAL
}

func NewNoop(d dal.DAL) NoteCache {
	return noop{
		DAL: d,
	}
}

func (n noop) Flush() error {
	return errors.New("noop cache: nothing to flush")
}
