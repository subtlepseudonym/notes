package dalpkg

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/subtlepseudonym/notes"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

const (
	defaultMetaFilename       = "meta"
	defaultNotesDirectory     = ".notes"
	defaultNoteFilenameFormat = "%06d"
)

// DAL interfaces the method by which we access the source of
// Meta and Note objects
type DAL interface {
	GetMeta() (*notes.Meta, error)
	SaveMeta(*notes.Meta) error

	GetNote(int) (*notes.Note, error)
	SaveNote(*notes.Note) error
	RemoveNote(int) error
}

type defaultDAL struct {
	version            string
	metaFilename       string
	notesDirectoryPath string
	noteFilenameFormat string
}

// NewDefaultDAL initializes a DAL with the default options
func NewDefaultDAL(version string) (DAL, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, errors.Wrap(err, "get home directory failed")
	}

	return &defaultDAL{
		version:            version,
		metaFilename:       defaultMetaFilename,
		notesDirectoryPath: path.Join(home, defaultNotesDirectory),
		noteFilenameFormat: defaultNoteFilenameFormat,
	}, nil
}

func (d *defaultDAL) buildNewMeta() (*notes.Meta, error) {
	if _, err := os.Stat(d.notesDirectoryPath); os.IsNotExist(err) {
		err = os.Mkdir(d.notesDirectoryPath, os.ModeDir|os.FileMode(0700))
		if err != nil {
			return nil, errors.Wrap(err, "create notes directory failed")
		}
	}

	metaPath := path.Join(d.notesDirectoryPath, d.metaFilename)
	metaFile, err := os.Create(metaPath)
	if err != nil {
		return nil, errors.Wrap(err, "create meta file failed")
	}
	defer metaFile.Close()

	m := &notes.Meta{
		Version: d.version,
		Notes:   make(map[int]notes.NoteMeta),
	}

	err = json.NewEncoder(metaFile).Encode(m)
	if err != nil {
		return nil, errors.Wrap(err, "encode meta file failed")
	}

	return m, errors.Wrap(metaFile.Close(), "close meta file failed")
}

// GetMeta retrieves and decodes a Meta from file
func (d *defaultDAL) GetMeta() (*notes.Meta, error) {
	metaPath := path.Join(d.notesDirectoryPath, d.metaFilename)
	metaFile, err := os.Open(metaPath)
	if err != nil {
		return d.buildNewMeta()
	}
	defer metaFile.Close()

	var m notes.Meta
	err = json.NewDecoder(metaFile).Decode(&m)
	if err != nil {
		return nil, errors.Wrap(err, "decode meta file failed")
	}

	return &m, errors.Wrap(metaFile.Close(), "close meta file failed")
}

// SaveMeta encodes and saves the provided Meta to file
func (d *defaultDAL) SaveMeta(meta *notes.Meta) error {
	metaPath := path.Join(d.notesDirectoryPath, d.metaFilename)
	err := os.Rename(metaPath, metaPath+".bak")
	if err != nil {
		return errors.Wrap(err, "backup old meta file failed")
	}
	// TODO: remove meta backup

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
		return errors.Wrap(err, "encode meta file failed")
	}

	return errors.Wrap(metaFile.Close(), "close meta file failed")
}

func (d *defaultDAL) getNotePath(id int) string {
	noteFilename := fmt.Sprintf(d.noteFilenameFormat, id)
	return path.Join(d.notesDirectoryPath, noteFilename)
}

// GetNote retrieves and decodes a Note from file
func (d *defaultDAL) GetNote(id int) (*notes.Note, error) {
	notePath := d.getNotePath(id)
	noteFile, err := os.Open(notePath)
	if err != nil {
		return nil, errors.Wrap(err, "open note file failed")
	}
	defer noteFile.Close()

	var n notes.Note
	err = json.NewDecoder(noteFile).Decode(&n)
	if err != nil {
		return nil, errors.Wrap(err, "decode note file failed")
	}

	return &n, errors.Wrap(noteFile.Close(), "close note file failed")
}

// SaveNote encodes and saves the provided Note to file
func (d *defaultDAL) SaveNote(note *notes.Note) error {
	notePath := d.getNotePath(note.Meta.ID)
	noteFile, err := os.Create(notePath)
	if err != nil {
		return errors.Wrap(err, "create note file failed")
	}
	defer noteFile.Close()

	err = json.NewEncoder(noteFile).Encode(note)
	if err != nil {
		os.Remove(notePath) // FIXME: do something with this error
		return errors.Wrap(err, "encode note file failed")
	}

	return errors.Wrap(noteFile.Close(), "close note file failed")
}

// RemoveNote deletes the note file
func (d *defaultDAL) RemoveNote(id int) error {
	notePath := d.getNotePath(id)
	err := os.Remove(notePath)
	if err != nil {
		return errors.Wrap(err, "remove note file failed")
	}

	return nil
}
