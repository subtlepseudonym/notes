package operations

import (
	"fmt"
	"time"

	"github.com/subtlepseudonym/notes"

	"go.uber.org/zap"
)

const (
	defaultDateTitleFormat   = time.RFC1123
	defaultDateTitleLocation = "UTC"
)

// NewNoteOptions provides values by which to alter the Note created by NewNote
type NewNoteOptions struct {
	Title        string `json:"title"`
	DateFormat   string `json:"dateFormat"`
	DateLocation string `json:"dateLocation"`
}

// NewNote creates a new note object according to the provided options and populates
// the body with the provided UpdateBodyFunc
func NewNote(ctx *Context, options NewNoteOptions) (*Context, error) {
	newNoteID := ctx.Meta.LatestID + 1
	if _, err := ctx.DAL.GetNoteMeta(newNoteID); err == nil {
		return ctx, fmt.Errorf("note ID %d (%x) already exists", newNoteID, newNoteID)
	}

	title := options.Title
	if title == "" {
		title = timestampTitle(ctx, options.DateFormat, options.DateLocation)
	}

	note := &notes.Note{
		Meta: notes.NoteMeta{
			ID:      newNoteID,
			Title:   title,
			Created: notes.JSONTime{time.Now()},
			Deleted: notes.JSONTime{time.Unix(0, 0)},
		},
	}

	err := ctx.DAL.SaveNote(note)
	if err != nil {
		return ctx, fmt.Errorf("save note: %v", err)
	}
	ctx.Logger.Debug(
		"created new note",
		zap.Int("noteID", note.Meta.ID),
		zap.String("notebook", ctx.DAL.GetNotebook()),
	)

	ctx.Meta.LatestID = note.Meta.ID
	metaSize, err := ctx.Meta.ApproxSize()
	if err != nil {
		ctx.Logger.Error("failed to approximate meta size", zap.Error(err))
	} else {
		ctx.Meta.Size = metaSize
	}

	err = ctx.DAL.SaveMeta(ctx.Meta)
	if err != nil {
		return ctx, fmt.Errorf("save meta: %v", err)
	}

	return ctx, nil
}

func timestampTitle(ctx *Context, format, location string) string {
	if format == "" {
		format = defaultDateTitleFormat
	}

	if location == "" {
		location = defaultDateTitleLocation
	}
	loc, err := time.LoadLocation(location)
	if err != nil {
		loc = time.UTC
		ctx.Logger.Error("failed to load location, defaulting to UTC", zap.Error(err), zap.String("location", location))
	}

	return time.Now().In(loc).Format(format)
}
