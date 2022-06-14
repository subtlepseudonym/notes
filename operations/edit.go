package operations

import (
	"time"

	"github.com/subtlepseudonym/notes"
	"github.com/subtlepseudonym/notes/dal"
)

// EditNote updates a note, defined by the given ID, as updates arrive on the given channel
func EditNote(data dal.DAL, note *notes.Note) (*notes.Note, error) {
	note.UpdatedAt = time.Now()

	err := data.WriteNote(note)
	if err != nil {
		return note, fmt.Errorf("write note: %w", err)
	}
}
