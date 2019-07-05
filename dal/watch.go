package dalpkg

import (
	"io/ioutil"
	"time"

	"github.com/pkg/errors"
	"github.com/subtlepseudonym/notes"
	"go.uber.org/zap"
)

// WatchAndUpdate periodically reads the contents of the provided file and compares
// it to the body of the provided note. If they aren't equal, it saves the changes
// to the DAL
func WatchAndUpdate(dal DAL, meta *notes.Meta, note *notes.Note, filename string, period time.Duration, stop chan struct{}, log *zap.Logger) error {
	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return nil
		case timestamp := <-ticker.C:
			b, err := ioutil.ReadFile(filename)
			if err != nil {
				return errors.Wrap(err, "read file failed") // FIXME: might want to log these rather than returning
			}

			fileContents := string(b)
			if note.Body == fileContents {
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
			log.Info("note updated", zap.Int("noteID", note.Meta.ID))

			metaSize, err := meta.ApproxSize()
			if err != nil {
				return errors.Wrap(err, "get meta size failed")
			}

			meta.Size = metaSize
			meta.Notes[note.Meta.ID] = note.Meta
			err = dal.SaveMeta(meta)
			if err != nil {
				return errors.Wrap(err, "save meta failed")
			}
			log.Info("meta updated", zap.Int("metaSize", meta.Size))
		}
	}

	return nil
}
