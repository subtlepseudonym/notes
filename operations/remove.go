package operations

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

type RemoveNoteOptions struct {
	HardDelete bool `json:"hardDelete"`
}

func RemoveNote(ctx *Context, options RemoveNoteOptions, noteID int) (*Context, error) {
	if options.HardDelete {
		err := ctx.DAL.RemoveNote(noteID)
		if err != nil {
			return ctx, fmt.Errorf("delete note: %v", err)
		}

		ctx.Logger.Debug("deleted note", zap.Int("noteID", noteID))
		return ctx, nil
	}

	note, err := ctx.DAL.GetNote(noteID)
	if err != nil {
		return ctx, fmt.Errorf("get note: %v", err)
	}

	note.Meta.Deleted.Time = time.Now()
	err = ctx.DAL.SaveNote(note)
	if err != nil {
		return ctx, fmt.Errorf("save note: %v", err)
	}

	ctx.Logger.Debug("soft-deleted note", zap.Int("noteID", noteID))
	return ctx, nil
}
