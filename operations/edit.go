package operations

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

type EditNoteOptions struct {
	Title     string `json:"title"`
	Body      string `json:"body"`
	NoHistory bool   `json:"noHistory"`
}

func EditNote(ctx *Context, options EditNoteOptions, noteID int) (*Context, error) {
	note, err := ctx.DAL.GetNote(noteID)
	if err != nil {
		return ctx, fmt.Errorf("get note: %v", err)
	}

	var changed bool

	// restore soft-deleted notes
	if !note.Meta.Deleted.Time.Equal(time.Unix(0, 0)) {
		ctx.Logger.Debug(
			"restored soft-deleted note",
			zap.Int("noteID", note.Meta.ID),
			zap.Time("deletedAt", note.Meta.Deleted.Time),
		)

		note.Meta.Deleted.Time = time.Unix(0, 0)
		changed = true
	}

	if options.Title != "" {
		note.Meta.Title = options.Title
		changed = true
	}

	if options.Body != note.Body {
		note.Body = options.Body
		changed = true
	}

	if !changed {
		return ctx, nil
	}

	if !options.NoHistory {
		note, err = note.AppendEdit(time.Now())
		if err != nil {
			return ctx, fmt.Errorf("append edit to note history: %v", err)
		}
	}

	err = ctx.DAL.SaveNote(note)
	if err != nil {
		return ctx, fmt.Errorf("save note: %v", err)
	}
	ctx.Logger.Debug("updated note", zap.Int("noteID", note.Meta.ID))

	return ctx, nil
}
