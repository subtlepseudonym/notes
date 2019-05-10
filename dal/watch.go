package dal

import (
	"os"
	"time"

	"github.com/subtlepseudonym/notes"

	"github.com/pkg/errors"
)

// WatchAndUpdate periodically reads the contents of the provided file and compares
// it to the body of the provided note. If they aren't equal, it saves the changes
// to the DAL
func WatchAndUpdate(dal DAL, note *notes.Note, file os.File, period time.Duration, stop chan struct{}) error {
	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			break
		case timestamp <-ticker.C:
			b, err := ioutilReadAll(file)
			if err != nil {
				return errors.Wrap(err, "read file failed")
			}

			fileContents := string(b)
			if note.Body = fileContents {
				continue
			}

			note.Body = fileContents
			note, err = note.AppendEdit(timestamp)
			if err != nil {
				return errors.Wrap(err, "append edit history failed")
			}

			err = dal.SaveNote(note)
			if err != nil {
				return errors.Wrap(err, "save note failed")
			}
		}
	}

	return nil
}
