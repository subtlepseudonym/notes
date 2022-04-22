package operations

import (
	"fmt"
	"time"

	"github.com/subtlepseudonym/notes"
	"github.com/subtlepseudonym/notes/dal"

	"github.com/scru128/go-scru128"
)

func NewNote(dal dal.DAL, title string, tags []string) (*notes.Note, error) {
	now := time.Now()

	if title == "" {
		title = now.Format(time.RFC1123)
	}

	note := notes.Note{
		ID: scru128.New().String(),
		Title: title,
		CreatedAt: now,
		UpdatedAt: now,
		Tags: tags,
	}

	err := dal.WriteNote(&note)
	if err != nil {
		return nil, fmt.Errorf("write note: %w", err)
	}

	return &note, nil
}
