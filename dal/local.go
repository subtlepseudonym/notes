package dal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"sync"

	"github.com/subtlepseudonym/notes"
)

const (
	defaultIndexFilename      = "index"
	defaultMetaFilename       = "meta"
	defaultNoteFilenameFormat = "%06d"
	noteFilenameRegex         = `[0-9]{6}`
)

type local struct {
	sync.Mutex
	version            string
	indexFilename      string
	metaFilename       string
	notesDirectoryPath string
	noteFilenameFormat string
}

// NewLocalDAL initializes a DAL with the default options
func NewLocalDAL(dirName, version string) (DAL, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home directory: %v", err)
	}

	return &local{
		version:            version,
		indexFilename:      defaultIndexFilename,
		metaFilename:       defaultMetaFilename,
		notesDirectoryPath: path.Join(home, dirName),
		noteFilenameFormat: defaultNoteFilenameFormat,
	}, nil
}

func (d *local) buildNewIndex() (notes.Index, error) {
	if _, err := os.Stat(d.notesDirectoryPath); os.IsNotExist(err) {
		err = os.Mkdir(d.notesDirectoryPath, os.ModeDir|os.FileMode(0700))
		if err != nil {
			return nil, fmt.Errorf("create notes directory: %v", err)
		}
	}

	indexPath := path.Join(d.notesDirectoryPath, d.indexFilename)
	indexFile, err := os.Create(indexPath)
	if err != nil {
		return nil, fmt.Errorf("create index file: %v", err)
	}
	defer indexFile.Close()

	index := notes.NewIndex(0) // use default capacity
	infos, err := ioutil.ReadDir(d.notesDirectoryPath)
	if err != nil {
		return nil, fmt.Errorf("read notes directory: %v", err)
	}

	nameRegex := regexp.MustCompile(noteFilenameRegex)
	for _, info := range infos {
		if info.IsDir() || !nameRegex.MatchString(info.Name()) {
			continue
		}

		noteFilename := path.Join(d.notesDirectoryPath, info.Name())
		noteFile, err := os.Open(noteFilename)
		if err != nil {
			// TODO: log this
			continue
		}

		var note notes.Note
		err = json.NewDecoder(noteFile).Decode(&note)
		if err != nil {
			// TODO: log this
			continue
		}

		index[note.Meta.ID] = note.Meta
	}

	err = json.NewEncoder(indexFile).Encode(index)
	if err != nil {
		return nil, fmt.Errorf("encode index file: %v", err)
	}

	err = indexFile.Close()
	if err != nil {
		return index, fmt.Errorf("close index file: %v", err)
	}
	return index, nil
}

func (d *local) GetIndex() (notes.Index, error) {
	d.Lock()
	defer d.Unlock()

	indexPath := path.Join(d.notesDirectoryPath, d.indexFilename)
	indexFile, err := os.Open(indexPath)
	if err != nil {
		return d.buildNewIndex()
	}
	defer indexFile.Close()

	var index notes.Index
	err = json.NewDecoder(indexFile).Decode(&index)
	if err != nil {
		return nil, fmt.Errorf("decode index file: %v", err)
	}

	err = indexFile.Close()
	if err != nil {
		return index, fmt.Errorf("close index file: %v", err)
	}
	return index, nil
}

func (d *local) SaveIndex(index notes.Index) error {
	d.Lock()
	defer d.Unlock()

	indexPath := path.Join(d.notesDirectoryPath, d.indexFilename)
	err := os.Rename(indexPath, indexPath+".bak")
	if err != nil {
		return fmt.Errorf("backup old index file: %v", err)
	}
	// TODO: remove index backup

	indexFile, err := os.Create(indexPath)
	if err != nil {
		err = os.Rename(indexPath+".bak", indexPath)
		if err != nil {
			return fmt.Errorf("restore index backup: %v", err)
		}
	}
	defer indexFile.Close()

	err = json.NewEncoder(indexFile).Encode(index)
	if err != nil {
		return fmt.Errorf("encode index file: %v", err)
	}

	err = indexFile.Close()
	if err != nil {
		return fmt.Errorf("close index file: %v", err)
	}
	return nil
}

func (d *local) buildNewMeta() (*notes.Meta, error) {
	if _, err := os.Stat(d.notesDirectoryPath); os.IsNotExist(err) {
		err = os.Mkdir(d.notesDirectoryPath, os.ModeDir|os.FileMode(0700))
		if err != nil {
			return nil, fmt.Errorf("create notes directory: %v", err)
		}
	}

	metaPath := path.Join(d.notesDirectoryPath, d.metaFilename)
	metaFile, err := os.Create(metaPath)
	if err != nil {
		return nil, fmt.Errorf("create meta file: %v", err)
	}
	defer metaFile.Close()

	m := &notes.Meta{
		Version: d.version,
	}

	err = json.NewEncoder(metaFile).Encode(m)
	if err != nil {
		return nil, fmt.Errorf("encode meta file: %v", err)
	}

	err = metaFile.Close()
	if err != nil {
		return m, fmt.Errorf("close meta file: %v", err)
	}
	return m, nil
}

// GetMeta retrieves and decodes a Meta from file
func (d *local) GetMeta() (*notes.Meta, error) {
	d.Lock()
	defer d.Unlock()

	metaPath := path.Join(d.notesDirectoryPath, d.metaFilename)
	metaFile, err := os.Open(metaPath)
	if err != nil {
		meta, err := d.buildNewMeta()
		return meta, err
	}
	defer metaFile.Close()

	var m notes.Meta
	err = json.NewDecoder(metaFile).Decode(&m)
	if err != nil {
		return nil, fmt.Errorf("decode meta file: %v", err)
	}

	err = metaFile.Close()
	if err != nil {
		return &m, fmt.Errorf("close meta file: %v", err)
	}
	return &m, nil
}

// SaveMeta encodes and saves the provided Meta to file
func (d *local) SaveMeta(meta *notes.Meta) error {
	d.Lock()
	defer d.Unlock()

	metaPath := path.Join(d.notesDirectoryPath, d.metaFilename)
	err := os.Rename(metaPath, metaPath+".bak")
	if err != nil {
		return fmt.Errorf("backup old meta file: %v", err)
	}
	// TODO: remove meta backup

	metaFile, err := os.Create(metaPath)
	if err != nil {
		err = os.Rename(metaPath+".bak", metaPath)
		if err != nil {
			return fmt.Errorf("restore meta backup: %v", err)
		}
		return fmt.Errorf("create meta file: %v", err)
	}
	defer metaFile.Close()

	err = json.NewEncoder(metaFile).Encode(meta)
	if err != nil {
		return fmt.Errorf("encode meta file: %v", err)
	}

	err = metaFile.Close()
	if err != nil {
		return fmt.Errorf("close meta file: %v", err)
	}
	return nil
}

func (d *local) getNotePath(id int) string {
	noteFilename := fmt.Sprintf(d.noteFilenameFormat, id)
	return path.Join(d.notesDirectoryPath, noteFilename)
}

// GetNote retrieves and decodes a Note from file
func (d *local) GetNote(id int) (*notes.Note, error) {
	notePath := d.getNotePath(id)
	noteFile, err := os.Open(notePath)
	if err != nil {
		return nil, fmt.Errorf("open note file: %v", err)
	}
	defer noteFile.Close()

	var n notes.Note
	err = json.NewDecoder(noteFile).Decode(&n)
	if err != nil {
		return nil, fmt.Errorf("decode note file: %v", err)
	}

	err = noteFile.Close()
	if err != nil {
		return &n, fmt.Errorf("close note file: %v", err)
	}
	return &n, nil
}

// SaveNote encodes and saves the provided Note to file
func (d *local) SaveNote(note *notes.Note) error {
	d.Lock()
	defer d.Unlock()

	notePath := d.getNotePath(note.Meta.ID)
	noteFile, err := os.Create(notePath)
	if err != nil {
		return fmt.Errorf("create note file: %v", err)
	}
	defer noteFile.Close()

	err = json.NewEncoder(noteFile).Encode(note)
	if err != nil {
		os.Remove(notePath) // FIXME: do something with this error
		return fmt.Errorf("encode note file: %v", err)
	}

	err = noteFile.Close()
	if err != nil {
		return fmt.Errorf("close note file: %v", err)
	}
	return nil
}

// RemoveNote deletes the note file
func (d *local) RemoveNote(id int) error {
	d.Lock()
	defer d.Unlock()

	notePath := d.getNotePath(id)
	err := os.Remove(notePath)
	if err != nil {
		return fmt.Errorf("remove note file: %v", err)
	}

	return nil
}
