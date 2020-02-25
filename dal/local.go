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
	defaultNotebook           = "notes"
	defaultIndexFilename      = "index"
	defaultMetaFilename       = "meta"
	defaultNoteFilenameFormat = "%06d"
	noteFilenameRegex         = `[0-9]{6}`
)

type local struct {
	sync.Mutex
	version            string
	baseDirectory      string
	notebook           string
	indexFilename      string
	metaFilename       string
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
		baseDirectory:      path.Join(home, dirName),
		notebook:           defaultNotebook,
		indexFilename:      defaultIndexFilename,
		metaFilename:       defaultMetaFilename,
		noteFilenameFormat: defaultNoteFilenameFormat,
	}, nil
}

func (d *local) notebookPath(name string) (string, error) {
	dirPath := path.Join(d.baseDirectory, d.notebook)
	info, err := os.Stat(dirPath)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		return dirPath, nil
	}
	return "", fmt.Errorf("notebook %q file exists, but is not a directory", dirPath)
}

func (d *local) CreateNotebook(name string) error {
	return nil
}

func (d *local) SetNotebook(name string) error {
	return nil
}

func (d *local) RemoveNotebook(name string) error {
	return nil
}

func (d *local) buildNewIndex() (notes.Index, error) {
	err := createDirIfNotExists(d.baseDirectory)
	if err != nil {
		return nil, fmt.Errorf("create base directory: %w", err)
	}

	notebookPath, err := d.notebookPath(d.notebook)
	if err != nil {
		return nil, fmt.Errorf("get notebook: %w", err)
	}

	index := notes.NewIndex(0) // use default capacity
	infos, err := ioutil.ReadDir(notebookPath)
	if err != nil {
		return nil, fmt.Errorf("read notes directory: %v", err)
	}

	nameRegex := regexp.MustCompile(noteFilenameRegex)
	for _, info := range infos {
		if info.IsDir() || !nameRegex.MatchString(info.Name()) {
			continue
		}

		noteFilename := path.Join(notebookPath, info.Name())
		notefile, err := os.Open(noteFilename)
		if err != nil {
			// todo: log this
			continue
		}

		var note notes.Note
		err = json.NewDecoder(notefile).Decode(&note)
		if err != nil {
			// todo: log this
			notefile.Close()
			continue
		}

		index[note.Meta.ID] = note.Meta
		notefile.Close()
	}

	indexPath := path.Join(notebookPath, d.indexFilename)
	indexFile, err := os.Create(indexPath)
	if err != nil {
		return nil, fmt.Errorf("create index file: %v", err)
	}

	err = json.NewEncoder(indexFile).Encode(index)
	if err != nil {
		indexFile.Close()
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

	notebookPath, err := d.notebookPath(d.notebook)
	if err != nil {
		return nil, fmt.Errorf("get notebook: %w", err)
	}

	indexPath := path.Join(notebookPath, d.indexFilename)
	indexFile, err := os.Open(indexPath)
	if err != nil {
		return d.buildNewIndex()
	}

	var index notes.Index
	err = json.NewDecoder(indexFile).Decode(&index)
	if err != nil {
		indexFile.Close()
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

	notebookPath, err := d.notebookPath(d.notebook)
	if err != nil {
		return fmt.Errorf("get notebook: %w", err)
	}

	indexPath := path.Join(notebookPath, d.indexFilename)
	err = os.Rename(indexPath, indexPath+".bak")
	if err != nil {
		return fmt.Errorf("backup index file: %v", err)
	}
	// TODO: remove index backup

	indexFile, err := os.Create(indexPath)
	if err != nil {
		err = os.Rename(indexPath+".bak", indexPath)
		if err != nil {
			return fmt.Errorf("restore index backup: %v", err)
		}
	}

	err = json.NewEncoder(indexFile).Encode(index)
	if err != nil {
		indexFile.Close()
		return fmt.Errorf("encode index file: %v", err)
	}

	err = indexFile.Close()
	if err != nil {
		return fmt.Errorf("close index file: %v", err)
	}
	return nil
}

func (d *local) buildNewMeta() (*notes.Meta, error) {
	err := createDirIfNotExists(d.baseDirectory)
	if err != nil {
		return nil, fmt.Errorf("create base directory: %w", err)
	}

	notebookPath, err := d.notebookPath(d.notebook)
	if err != nil {
		return nil, fmt.Errorf("get notebook: %w", err)
	}

	metaPath := path.Join(notebookPath, d.metaFilename)
	metaFile, err := os.Create(metaPath)
	if err != nil {
		return nil, fmt.Errorf("create meta file: %v", err)
	}

	m := &notes.Meta{
		Version: d.version,
	}

	err = json.NewEncoder(metaFile).Encode(m)
	if err != nil {
		metaFile.Close()
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

	notebookPath, err := d.notebookPath(d.notebook)
	if err != nil {
		return nil, fmt.Errorf("get notebook: %w", err)
	}

	metaPath := path.Join(notebookPath, d.metaFilename)
	metaFile, err := os.Open(metaPath)
	if err != nil {
		meta, err := d.buildNewMeta()
		return meta, err
	}

	var m notes.Meta
	err = json.NewDecoder(metaFile).Decode(&m)
	if err != nil {
		metaFile.Close()
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

	notebookPath, err := d.notebookPath(d.notebook)
	if err != nil {
		return fmt.Errorf("get notebook: %w", err)
	}

	metaPath := path.Join(notebookPath, d.metaFilename)
	err = os.Rename(metaPath, metaPath+".bak")
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

	err = json.NewEncoder(metaFile).Encode(meta)
	if err != nil {
		metaFile.Close()
		return fmt.Errorf("encode meta file: %v", err)
	}

	err = metaFile.Close()
	if err != nil {
		return fmt.Errorf("close meta file: %v", err)
	}
	return nil
}

func (d *local) getNotePath(id int) (string, error) {
	noteFilename := fmt.Sprintf(d.noteFilenameFormat, id)
	notebookPath, err := d.notebookPath(d.notebook)
	if err != nil {
		return "", fmt.Errorf("get notebook: %w", err)
	}

	return path.Join(notebookPath, noteFilename), nil
}

// GetNote retrieves and decodes a Note from file
func (d *local) GetNote(id int) (*notes.Note, error) {
	notePath, err := d.getNotePath(id)
	if err != nil {
		return nil, fmt.Errorf("get note path: %v", err)
	}

	noteFile, err := os.Open(notePath)
	if err != nil {
		return nil, fmt.Errorf("open note file: %v", err)
	}

	var n notes.Note
	err = json.NewDecoder(noteFile).Decode(&n)
	if err != nil {
		noteFile.Close()
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

	notePath, err := d.getNotePath(note.Meta.ID)
	if err != nil {
		return fmt.Errorf("get note path: %v", err)
	}

	noteFile, err := os.Create(notePath)
	if err != nil {
		return fmt.Errorf("create note file: %v", err)
	}

	err = json.NewEncoder(noteFile).Encode(note)
	if err != nil {
		noteFile.Close()
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

	notePath, err := d.getNotePath(id)
	if err != nil {
		return fmt.Errorf("get note path: %v", err)
	}

	err = os.Remove(notePath)
	if err != nil {
		return fmt.Errorf("remove note file: %v", err)
	}

	return nil
}

func createDirIfNotExists(dirname string) error {
	info, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		return os.Mkdir(dirname, os.ModeDir|os.FileMode(0700))
	}

	if info.IsDir() {
		return nil
	}
	return fmt.Errorf("file %q exists, but is not a directory", dirname)
}
