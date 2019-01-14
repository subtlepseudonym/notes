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
	defaultMetaFilename       = "meta"
	defaultNotesDir           = ".notes"
	defaultEditor             = "vim"
	defaultNoteFilenameFormat = "%06d"
)

// Meta holds meta information for the local notes storage as a whole
type Meta struct {
	Version  string           `json:"version"`
	LatestID int              `json:"latestId"`
	Notes    map[int]NoteMeta `json:"notes"` // maps note ID to NoteMeta
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
func buildNewMeta(version string) (Meta, error) {
	notesDir, err := getNotesDirPath()
	if err != nil {
		return Meta{}, errors.Wrap(err, "get meta dir failed")
	}

	if _, err = os.Stat(notesDir); os.IsNotExist(err) {
		err = os.Mkdir(notesDir, os.ModeDir|os.FileMode(0700))
		if err != nil {
			return Meta{}, errors.Wrap(err, "create notes directory failed")
		}
	}

	metaPath := path.Join(notesDir, defaultMetaFilename)
	f, err := os.Create(metaPath)
	if err != nil {
		return Meta{}, errors.Wrap(err, "create meta file failed")
	}

	m := Meta{
		Version: version,
		Notes:   make(map[int]NoteMeta),
	}

	encoder := json.NewEncoder(f)
	err = encoder.Encode(&m)
	if err != nil {
		return Meta{}, errors.Wrap(err, "encode meta object failed")
	}

	// TODO: log that this function was called
	return m, nil
}

// GetMeta retrieves global meta info from file
func GetMeta(version string) (Meta, error) {
	notesDir, err := getNotesDirPath()
	if err != nil {
		return Meta{}, errors.Wrap(err, "get meta dir failed")
	}

	// TODO: decide if we care about customizing the meta file name
	metaPath := path.Join(notesDir, defaultMetaFilename) // FIXME: don't use default
	f, err := os.Open(metaPath)
	if err != nil {
		return buildNewMeta(version)
	}

	decoder := json.NewDecoder(f)

	var m Meta
	err = decoder.Decode(&m)
	if err != nil {
		return Meta{}, errors.Wrap(err, "decode meta object failed")
	}

	return m, nil
}

// SaveMeta saves the provided meta to file
func SaveMeta(meta Meta) error {
	notesDir, err := getNotesDirPath()
	if err != nil {
		return errors.Wrap(err, "get meta dir failed")
	}

	metaPath := path.Join(notesDir, defaultMetaFilename) // FIXME: don't use default
	err = os.Rename(metaPath, metaPath+".bak")
	if err != nil {
		return errors.Wrap(err, "backup old meta failed")
	}

	metaFile, err := os.Create(metaPath)
	if err != nil {
		err = os.Rename(metaPath+".bak", metaPath)
		if err != nil {
			return errors.Wrap(err, "restoring meta backup failed")
		}
		return errors.Wrap(err, "create meta file failed")
	}
	defer metaFile.Close()

	err = json.NewEncoder(metaFile).Encode(meta)
	if err != nil {
		os.Remove(metaPath) // FIXME: this could return a path error
		return errors.Wrap(err, "encode meta object failed")
	}

	return errors.Wrap(metaFile.Close(), "close meta file failed")
}

func getNoteFilename(id int) (string, error) {
	notesDir, err := getNotesDirPath()
	if err != nil {
		return "", errors.Wrap(err, "get notes dir failed")
	}

	// FIXME: don't use default note filename format
	// TODO: can we validate a provided string format?
	return path.Join(notesDir, fmt.Sprintf(defaultNoteFilenameFormat, id)), nil
}

// GetNote retrieves a note from file by ID
func GetNote(id int) (Note, error) {
	notePath, err := getNoteFilename(id)
	if err != nil {
		return Note{}, errors.Wrap(err, "get note filename failed")
	}

	f, err := os.Open(notePath)
	if err != nil {
		return Note{}, errors.Wrap(err, "open note file failed")
	}
	defer f.Close()

	var n Note
	err = json.NewDecoder(f).Decode(&n)
	if err != nil {
		return Note{}, errors.Wrap(err, "decode note object failed")
	}

	return n, errors.Wrap(f.Close(), "close note file failed")
}

// SaveNote saves the provided note to file
func SaveNote(note Note) error {
	notePath, err := getNoteFilename(note.Meta.ID)
	if err != nil {
		return errors.Wrap(err, "get note filename failed")
	}

	noteFile, err := os.Create(notePath)
	if err != nil {
		return errors.Wrap(err, "create note file failed")
	}

	err = json.NewEncoder(noteFile).Encode(note)
	if err != nil {
		os.Remove(notePath) // FIXME: this could return a path error (extremely unlikely though)
		return errors.Wrap(err, "encode note object failed")
	}

	return nil
}

// RemoveNote removes a note from disk
func RemoveNote(id int) error {
	notePath, err := getNoteFilename(id)
	if err != nil {
		return errors.Wrap(err, "get note filename failed")
	}

	err = os.Remove(notePath)
	if err != nil {
		return errors.Wrap(err, "remove note file failed")
	}

	return nil
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
