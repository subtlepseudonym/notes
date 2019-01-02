package files

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

const (
	defaultMetaFilename = "meta"
	defaultNotesDir     = ".notes"
	defaultEditor       = "vim"
)

// meta holds meta information for the local notes storage as a whole
type meta struct {
	Version string     `json:"version"`
	Notes   []NoteMeta `json:"notes"`
}

// NoteMeta holds meta information for one note to make commands that only access
// meta information perform faster
type NoteMeta struct {
	ID      int       `json:"id"` // incremented starting at 1
	Title   string    `json:"title"`
	Created time.Time `json:"created"`
	Deleted time.Time `json:"deleted"`
}

// Note includes the content of the note as well as its meta information as backup in
// case we need to recreate the meta file from scratch
type Note struct {
	Meta NoteMeta `json:"meta"`
	Body string   `json:"body"`
}

// getNotesDirPath returns the path to the default directory containing notes files
func getNotesDirPath() (string, error) { // FIXME: don't use default
	home, err := homedir.Dir()
	if err != nil {
		return "", errors.Wrap(err, "get home directory failed")
	}

	return path.Join(home, defaultNotesDir), nil
}

// buildNewMeta creates a new, empty meta object with only the Version field
// specified and writes it to the notes directory
func buildNewMeta(version string) (meta, error) {
	notesDir, err := getNotesDirPath()
	if err != nil {
		return meta{}, errors.Wrap(err, "get meta dir failed")
	}

	if _, err = os.Stat(notesDir); os.IsNotExist(err) {
		err = os.Mkdir(notesDir, os.ModeDir|os.FileMode(0700))
		if err != nil {
			return meta{}, errors.Wrap(err, "create notes directory failed")
		}
	}

	metaPath := path.Join(notesDir, defaultMetaFilename)
	f, err := os.Create(metaPath)
	if err != nil {
		return meta{}, errors.Wrap(err, "create meta file failed")
	}

	encoder := json.NewEncoder(f)

	m := meta{
		Version: version,
	}
	err = encoder.Encode(&m)
	if err != nil {
		return meta{}, errors.Wrap(err, "encode meta object failed")
	}

	// TODO: log that this function was called
	return m, nil
}

// GetMeta retrieves global meta info from file
func GetMeta(version string) (meta, error) {
	notesDir, err := getNotesDirPath()
	if err != nil {
		return meta{}, errors.Wrap(err, "get meta dir failed")
	}

	metaPath := path.Join(notesDir, defaultMetaFilename) // FIXME: don't use default
	f, err := os.Open(metaPath)
	if err != nil { // TODO: create notes dir / meta file here rather than in lsBeforeFunc(...)
		return buildNewMeta(version)
		// return meta{}, errors.Wrap(err, "open meta file failed")
	}

	decoder := json.NewDecoder(f)

	var m meta
	err = decoder.Decode(&m)
	if err != nil {
		return meta{}, errors.Wrap(err, "decode meta object failed")
	}

	return m, nil
}

// GetNote retrieves a note from file by ID
func GetNote(id int) (Note, error) {
	notesDir, err := getNotesDirPath()
	if err != nil {
		return Note{}, errors.Wrap(err, "get notes dir failed")
	}

	notePath := path.Join(notesDir, fmt.Sprintf("%06d", id))
	f, err := os.Open(notePath)
	if err != nil {
		return Note{}, errors.Wrap(err, "open note file failed")
	}

	var n Note
	err = json.NewDecoder(f).Decode(&n)
	if err != nil {
		return Note{}, errors.Wrap(err, "decode note object failed")
	}

	return n, nil
}

// SaveNote saves the provided note to file
func SaveNote(note Note) error {
	notesDir, err := getNotesDirPath()
	if err != nil {
		return errors.Wrap(err, "get notes dir failed")
	}

	notePath := path.Join(notesDir, fmt.Sprintf("%06d", note.Meta.ID))
	noteFile, err := os.Create(notePath)
	if err != nil {
		return errors.Wrap(err, "create note file failed")
	}

	err = json.NewEncoder(noteFile).Encode(note)
	if err != nil {
		return errors.Wrap(err, "encode note object failed")
	}

	return nil
}

// GetNoteBodyFromUser drops the user into the provided editor command before
// retrieving the contents of the edited file
func GetNoteBodyFromUser(editor string) (string, error) {
	tempFile, err := ioutil.TempFile("", "note")
	if err != nil {
		return "", errors.Wrap(err, "create temporary file failed")
	}
	defer tempFile.Close()

	cmd := exec.Command(editor, tempFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	err = cmd.Run()
	if err != nil {
		return "", errors.Wrap(err, "run editor command failed")
	}

	bodyBytes, err := ioutil.ReadAll(tempFile)
	if err != nil {
		return "", errors.Wrap(err, "read temporary file failed")
	}

	return string(bodyBytes), errors.Wrap(tempFile.Close(), "close temporary file failed")
}
