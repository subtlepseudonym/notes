package notes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/pkg/errors"
)

// Meta holds meta information for the local notes storage as a whole
type Meta struct {
	Version  string           `json:"version"`
	LatestID int              `json:"latestId"`
	Notes    map[int]NoteMeta `json:"notes"` // maps note ID to NoteMeta
}

type JSONTime struct {
	time.Time
}

func (j JSONTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.UnixNano()) // FIXME: this will fail after 2262
}

func (j *JSONTime) UnmarshalJSON(b []byte) error {
	var i int64
	err := json.Unmarshal(b, &i)
	if err != nil {
		return err
	}

	j.Time = time.Unix(0, i)
	return nil
}

// NoteMeta holds meta information for one note to make commands that only access
// meta information perform faster
type NoteMeta struct {
	ID      int        `json:"id"` // incremented starting at 1
	Title   string     `json:"title"`
	Created JSONTime   `json:"created"`
	Updated []JSONTime `json:"updated"`
	Deleted JSONTime   `json:"deleted"`
}

// Note includes the content of the note as well as its meta information as backup in
// case we need to recreate the meta file from scratch
type Note struct {
	Meta NoteMeta `json:"meta"`
	Body string   `json:"body"`
}

// GetNoteBodyFromUser drops the user into the provided editor command before
// retrieving the contents of the edited file
func GetNoteBodyFromUser(editor, existingBody string) (string, error) {
	file, err := ioutil.TempFile("", "note")
	if err != nil {
		return "", errors.Wrap(err, "create temporary file failed")
	}
	defer file.Close()

	_, err = fmt.Fprint(file, existingBody)
	if err != nil {
		return "", errors.Wrap(err, "print existing body to temporary file failed")
	}

	cmd := exec.Command(editor, file.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	err = cmd.Run()
	if err != nil {
		return "", errors.Wrap(err, "run editor command failed")
	}

	bodyBytes, err := ioutil.ReadFile(file.Name())
	if err != nil {
		return "", errors.Wrap(err, "read temporary file failed")
	}

	return string(bodyBytes), errors.Wrap(file.Close(), "close temporary file failed")
}
