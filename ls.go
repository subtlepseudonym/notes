package notes

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/subtlepseudonym/notes/files"

	"github.com/pkg/errors"
)

const (
	defaultListSize       = 10
	defaultListTimeFormat = time.RFC822
	defaultListDelimiter  = "|"
)

// ListOptions defines the set of options for modifying the behavior
// of the ls command
type ListOptions struct {
	ShowAll     bool
	LongFormat  bool
	ShowDeleted bool

	ListSize      int
	TimeFormat    string
	ListDelimiter string
}

func List(output io.Writer, options ListOptions) error {
	meta, err := files.GetMeta(Version)
	if err != nil {
		return errors.Wrap(err, "get meta failed")
	}

	limit := defaultListSize
	if options.ShowAll {
		limit = len(meta.Notes)
	} else if options.ListSize != 0 {
		limit = options.ListSize
	}

	idFormat := fmt.Sprintf("%% %dx", len(meta.Notes)+1)

	var fields []string
	var listed int
	idx := len(meta.Notes) - 1

	for listed < limit && idx >= 0 {
		note := meta.Notes[idx]

		fields = append(fields, fmt.Sprintf(idFormat, note.ID))

		if options.ShowDeleted {
			if time.Unix(0, 0).UTC().Equal(note.Deleted.UTC()) {
				fields = append(fields, " ")
			} else {
				fields = append(fields, "d")
			}
		}

		if options.LongFormat {
			timeFormat := defaultListTimeFormat
			if options.TimeFormat != "" {
				timeFormat = options.TimeFormat
			}
			fields = append(fields, note.Created.Format(timeFormat))
		}

		fields = append(fields, note.Title)

		delimiter := defaultListDelimiter
		if options.ListDelimiter != "" {
			delimiter = options.ListDelimiter
		}
		fmt.Fprintln(output, strings.Join(fields, delimiter))
	}

	return nil
}
