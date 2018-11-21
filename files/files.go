package files

import (
	"encoding/gob"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

const (
	defaultNotesDir     = ".notes"
	defaultMetaFilename = "meta"
)

// meta holds meta information for the local notes storage as a whole
type meta struct {
	Version string
	Notes   []NoteMeta
}

// NoteMeta holds meta information for one note to make commands that only access
// meta information perform faster
type NoteMeta struct {
	ID      int // incremented starting at 1
	Title   string
	Created time.Time
	Deleted time.Time
}

// Note includes the content of the note as well as its meta information as backup in
// case we need to recreate the meta file from scratch
type Note struct {
	Meta NoteMeta
	Body string
}

// getNotesDir returns the path to the default directory containing notes files
func getNotesDir() (string, error) { // FIXME: don't use default
	home, err := homedir.Dir()
	if err != nil {
		return "", errors.Wrap(err, "get home directory failed")
	}

	return path.Join(home, defaultNotesDir), nil
}

// GetMeta reads the meta object from the notes directory and returns it
func GetMeta() (meta, error) {
	notesDir, err := getNotesDir()
	if err != nil {
		return meta{}, errors.Wrap(err, "get meta dir failed")
	}

	metaPath := path.Join(notesDir, defaultMetaFilename) // FIXME: don't use default
	f, err := os.Open(metaPath)
	if err != nil {
		return meta{}, errors.Wrap(err, "open meta file failed")
	}

	decoder := gob.NewDecoder(f)

	var m meta
	err = decoder.Decode(&m)
	if err != nil {
		return meta{}, errors.Wrap(err, "decode meta object failed")
	}

	return m, nil
}

// GetNote reads a Note struct from the notes directory given by the id argument
func GetNote(id int) (Note, error) {
	notesDir, err := getNotesDir()
	if err != nil {
		return Note{}, errors.Wrap(err, "get notes dir failed")
	}

	notePath := path.Join(notesDir, fmt.Sprintf("%06d", id))
	f, err := os.Open(notePath)
	if err != nil {
		return Note{}, errors.Wrap(err, "open note file failed")
	}

	decoder := gob.NewDecoder(f)

	var n Note
	err = decoder.Decode(&n)
	if err != nil {
		return Note{}, errors.Wrap(err, "decode note object failed")
	}

	return n, nil
}
