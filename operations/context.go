package operations

import (
	"github.com/subtlepseudonym/notes"
	"github.com/subtlepseudonym/notes/dal"

	"go.uber.org/zap"
)

type Context struct {
	Meta *notes.Meta
	DAL  dal.DAL

	Logger *zap.Logger
}

type UpdateBodyFunc func() (string, error)
