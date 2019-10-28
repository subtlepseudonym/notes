package dalpkg

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/subtlepseudonym/notes"

	"github.com/mitchellh/go-homedir"
)

const (
	defaultMetaFilename       = "meta"
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

type localDAL struct {
	version            string
	metaFilename       string
	notesDirectoryPath string
	noteFilenameFormat string
}

// NewLocalDAL initializes a DAL with the default options
func NewLocalDAL(dirName, version string) (DAL, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("get home directory: %w", err)
	}

	return &localDAL{
		version:            version,
		metaFilename:       defaultMetaFilename,
		notesDirectoryPath: path.Join(home, dirName),
		noteFilenameFormat: defaultNoteFilenameFormat,
	}, nil
}

func (d *localDAL) buildNewMeta() (*notes.Meta, error) {
	if _, err := os.Stat(d.notesDirectoryPath); os.IsNotExist(err) {
		err = os.Mkdir(d.notesDirectoryPath, os.ModeDir|os.FileMode(0700))
		if err != nil {
			return nil, fmt.Errorf("create notes directory: %w", err)
		}
	}

	metaPath := path.Join(d.notesDirectoryPath, d.metaFilename)
	metaFile, err := os.Create(metaPath)
	if err != nil {
		return nil, fmt.Errorf("create meta file: %w", err)
	}
	defer metaFile.Close()

	m := &notes.Meta{
		Version: d.version,
		Notes:   make(map[int]notes.NoteMeta),
	}

	err = json.NewEncoder(metaFile).Encode(m)
	if err != nil {
		return nil, fmt.Errorf("encode meta file: %w", err)
	}

	err = metaFile.Close()
	if err != nil {
		return m, fmt.Errorf("close meta file: %w", err)
	}
	return m, nil
}

// GetMeta retrieves and decodes a Meta from file
func (d *localDAL) GetMeta() (*notes.Meta, error) {
	metaPath := path.Join(d.notesDirectoryPath, d.metaFilename)
	metaFile, err := os.Open(metaPath)
	if err != nil {
		return d.buildNewMeta()
	}
	defer metaFile.Close()

	var m notes.Meta
	err = json.NewDecoder(metaFile).Decode(&m)
	if err != nil {
		return nil, fmt.Errorf("decode meta file: %w", err)
	}

	err = metaFile.Close()
	if err != nil {
		return &m, fmt.Errorf("close meta file: %w", err)
	}
	return &m, nil
}

// SaveMeta encodes and saves the provided Meta to file
func (d *localDAL) SaveMeta(meta *notes.Meta) error {
	metaPath := path.Join(d.notesDirectoryPath, d.metaFilename)
	err := os.Rename(metaPath, metaPath+".bak")
	if err != nil {
		return fmt.Errorf("backup old meta file: %w", err)
	}
	// TODO: remove meta backup

	metaFile, err := os.Create(metaPath)
	if err != nil {
		err = os.Rename(metaPath+".bak", metaPath)
		if err != nil {
			return fmt.Errorf("restore meta backup: %w", err)
		}
		return fmt.Errorf("create meta file: %w", err)
	}
	defer metaFile.Close()

	err = json.NewEncoder(metaFile).Encode(meta)
	if err != nil {
		return fmt.Errorf("encode meta file: %w", err)
	}

	err = metaFile.Close()
	if err != nil {
		return fmt.Errorf("close meta file: %w", err)
	}
	return nil
}

func (d *localDAL) getNotePath(id int) string {
	noteFilename := fmt.Sprintf(d.noteFilenameFormat, id)
	return path.Join(d.notesDirectoryPath, noteFilename)
}

// GetNote retrieves and decodes a Note from file
func (d *localDAL) GetNote(id int) (*notes.Note, error) {
	notePath := d.getNotePath(id)
	noteFile, err := os.Open(notePath)
	if err != nil {
		return nil, fmt.Errorf("open note file: %w", err)
	}
	defer noteFile.Close()

	var n notes.Note
	err = json.NewDecoder(noteFile).Decode(&n)
	if err != nil {
		return nil, fmt.Errorf("decode note file: %w", err)
	}

	err = noteFile.Close()
	if err != nil {
		return &n, fmt.Errorf("close note file: %w", err)
	}
	return &n, nil
}

// SaveNote encodes and saves the provided Note to file
func (d *localDAL) SaveNote(note *notes.Note) error {
	notePath := d.getNotePath(note.Meta.ID)
	noteFile, err := os.Create(notePath)
	if err != nil {
		return fmt.Errorf("create note file: %w", err)
	}
	defer noteFile.Close()

	err = json.NewEncoder(noteFile).Encode(note)
	if err != nil {
		os.Remove(notePath) // FIXME: do something with this error
		return fmt.Errorf("encode note file: %w", err)
	}

	err = noteFile.Close()
	if err != nil {
		return fmt.Errorf("close note file: %w", err)
	}
	return nil
}

// RemoveNote deletes the note file
func (d *localDAL) RemoveNote(id int) error {
	notePath := d.getNotePath(id)
	err := os.Remove(notePath)
	if err != nil {
		return fmt.Errorf("remove note file: %w", err)
	}

	return nil
}
