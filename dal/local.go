package dal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"sync"

	"github.com/subtlepseudonym/notes"
)

const (
	defaultMetaFilename       = "meta"
	defaultIndexFilename      = "index"
	defaultNoteFilenameFormat = "%06d"
	noteFilenameRegex         = `[0-9]{6}`
	defaultIndexCapacity      = 256
)

type local struct {
	sync.Mutex
	baseDirectory      string
	notebook           string
	indexFilename      string
	metaFilename       string
	noteFilenameFormat string

	indexes map[string]map[int]notes.NoteMeta // map notebook name to map of IDs to NoteMeta
}

// NewLocal initializes a DAL with the default options
func NewLocal(dirName, version string) (DAL, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home directory: %v", err)
	}

	baseDirectory := path.Join(home, dirName)
	info, err := os.Stat(baseDirectory)
	if os.IsNotExist(err) {
		err = os.Mkdir(baseDirectory, os.ModeDir|os.FileMode(0700))
		if err != nil {
			return nil, fmt.Errorf("make base directory: %v", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("stat base directory: %v", err)
	} else if !info.IsDir() {
		return nil, fmt.Errorf("%s exists, but is not a directory", baseDirectory)
	}

	metaPath := path.Join(baseDirectory, defaultMetaFilename)
	_, err = os.Stat(metaPath)
	if os.IsNotExist(err) {
		err = buildMeta(baseDirectory, defaultMetaFilename, version)
		if err != nil {
			return nil, fmt.Errorf("build meta: %v", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("stat meta: %v", err)
	}

	var index map[int]notes.NoteMeta
	index, err = loadIndex(path.Join(baseDirectory, defaultIndexFilename))
	if errors.Is(err, os.ErrNotExist) {
		index, err = buildIndex(baseDirectory, "")
		if err != nil {
			return nil, fmt.Errorf("build index: %v", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("stat index: %v", err)
	}

	return &local{
		baseDirectory:      path.Join(home, dirName),
		metaFilename:       defaultMetaFilename,
		indexFilename:      defaultIndexFilename,
		noteFilenameFormat: defaultNoteFilenameFormat,
		indexes:            map[string]map[int]notes.NoteMeta{"": index},
	}, nil
}

// GetMeta retrieves and decodes a Meta from file
func (d *local) GetMeta() (*notes.Meta, error) {
	d.Lock()
	defer d.Unlock()

	metaPath := path.Join(d.baseDirectory, d.metaFilename)
	metaFile, err := os.Open(metaPath)
	if err != nil {
		return nil, fmt.Errorf("open meta file: %v", err)
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

	metaPath := path.Join(d.baseDirectory, d.metaFilename)
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

func (d *local) CreateNotebook(name string) error {
	notebookPath := path.Join(d.baseDirectory, name)
	err := os.Mkdir(notebookPath, os.ModeDir|os.FileMode(0700))
	if err != nil {
		return fmt.Errorf("make notebook directory: %v", err)
	}

	index, err := buildIndex(d.baseDirectory, name)
	if err != nil {
		return fmt.Errorf("build index: %v", err)
	}
	d.indexes[name] = index

	return nil
}

func (d *local) SetNotebook(name string) error {
	notebookPath := path.Join(d.baseDirectory, name)
	info, err := os.Stat(notebookPath)
	if err != nil {
		return fmt.Errorf("stat notebook directory: %v", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("file %s exists, but is not a directory", notebookPath)
	}

	d.Lock()
	d.notebook = name
	d.Unlock()

	return nil
}

func (d *local) RemoveNotebook(name string, recursive bool) error {
	notebookPath := path.Join(d.baseDirectory, name)
	info, err := os.Stat(notebookPath)
	if err != nil {
		return fmt.Errorf("stat notebook directory: %v", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("file %s exists, but is not a directory", notebookPath)
	}

	if recursive {
		return os.RemoveAll(notebookPath)
	}

	// TODO: check for notebook contents, remove index, then os.Remove
	return os.Remove(notebookPath)
}

func (d *local) GetNoteMeta(id int) (*notes.NoteMeta, error) {
	d.Lock()
	defer d.Unlock()

	index, ok := d.indexes[d.notebook]
	if !ok {
		return nil, fmt.Errorf("notebook %s index not found", d.notebook)
	}

	noteMeta, ok := index[id]
	if !ok {
		return nil, fmt.Errorf("note meta not in index")
	}
	return &noteMeta, nil
}

func (d *local) GetAllNoteMetas() (map[int]notes.NoteMeta, error) {
	d.Lock()
	defer d.Unlock()

	index, ok := d.indexes[d.notebook]
	if !ok {
		return nil, fmt.Errorf("notebook %s index not found", d.notebook)
	}

	return index, nil
}

func (d *local) getNotePath(id int) (string, error) {
	noteFilename := fmt.Sprintf(d.noteFilenameFormat, id)
	return path.Join(d.baseDirectory, d.notebook, noteFilename), nil
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

	index, ok := d.indexes[d.notebook]
	if !ok {
		return fmt.Errorf("notebook %s index not found", d.notebook)
	}
	index[note.Meta.ID] = note.Meta

	indexPath := path.Join(d.baseDirectory, d.notebook, defaultIndexFilename)
	err = saveIndex(indexPath, index)
	if err != nil {
		return fmt.Errorf("save index: %v", err)
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

	index, ok := d.indexes[d.notebook]
	if !ok {
		return fmt.Errorf("notebook %s index not found", d.notebook)
	}
	delete(index, id)

	indexPath := path.Join(d.baseDirectory, d.notebook, defaultIndexFilename)
	err = saveIndex(indexPath, index)
	if err != nil {
		return fmt.Errorf("save index: %v", err)
	}

	return nil
}

func createDirIfNotExists(dirname string) error {
	info, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		return os.Mkdir(dirname, os.ModeDir|os.FileMode(0700))
	}

	if !info.IsDir() {
		return fmt.Errorf("file %s exists, but is not a directory", dirname)
	}

	return nil
}

func buildMeta(baseDirectory, filename, version string) error {
	err := createDirIfNotExists(baseDirectory)
	if err != nil {
		return fmt.Errorf("create base directory: %w", err)
	}

	metaPath := path.Join(baseDirectory, filename)
	metaFile, err := os.Create(metaPath)
	if err != nil {
		return fmt.Errorf("create meta file: %v", err)
	}

	m := &notes.Meta{
		Version: version,
	}

	err = json.NewEncoder(metaFile).Encode(m)
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

func buildIndex(baseDirectory, notebook string) (map[int]notes.NoteMeta, error) {
	notebookPath := path.Join(baseDirectory, notebook)
	infos, err := ioutil.ReadDir(notebookPath)
	if err != nil {
		return nil, fmt.Errorf("read notes directory: %w", err)
	}

	index := make(map[int]notes.NoteMeta, defaultIndexCapacity)
	nameRegex := regexp.MustCompile(noteFilenameRegex)
	for _, info := range infos {
		if info.IsDir() || !nameRegex.MatchString(info.Name()) {
			continue
		}

		noteFilename := path.Join(notebookPath, info.Name())
		notefile, err := os.Open(noteFilename)
		if err != nil {
			// TODO: log this
			continue
		}

		var note notes.Note
		err = json.NewDecoder(notefile).Decode(&note)
		if err != nil {
			// TODO: log this
			notefile.Close()
			continue
		}

		index[note.Meta.ID] = note.Meta
		notefile.Close()
	}

	indexPath := path.Join(notebookPath, defaultIndexFilename)
	indexFile, err := os.Create(indexPath)
	if err != nil {
		return nil, fmt.Errorf("create index file: %w", err)
	}

	err = json.NewEncoder(indexFile).Encode(index)
	if err != nil {
		indexFile.Close()
		return nil, fmt.Errorf("encode index file: %w", err)
	}

	err = indexFile.Close()
	if err != nil {
		return nil, fmt.Errorf("close index file: %w", err)
	}
	return index, nil
}

func loadIndex(indexPath string) (map[int]notes.NoteMeta, error) {
	indexFile, err := os.Open(indexPath)
	if err != nil {
		return nil, fmt.Errorf("open index file: %w", err)
	}

	var index map[int]notes.NoteMeta
	err = json.NewDecoder(indexFile).Decode(&index)
	if err != nil {
		indexFile.Close()
		return nil, fmt.Errorf("decode index file: %w", err)
	}

	err = indexFile.Close()
	if err != nil {
		return nil, fmt.Errorf("close index file: %w", err)
	}
	return index, nil
}

// FIXME: saveIndex is called with default index path in SaveNote and
// RemoveNote, which will lead to clobbering when changing notebooks
func saveIndex(indexPath string, index map[int]notes.NoteMeta) error {
	err := os.Rename(indexPath, indexPath+".bak")
	if err != nil {
		return fmt.Errorf("backup index file: %w", err)
	}
	// TODO: remove index backup

	indexFile, err := os.Create(indexPath)
	if err != nil {
		err = os.Rename(indexPath+".bak", indexPath)
		if err != nil {
			return fmt.Errorf("restore index backup: %w", err)
		}
	}

	err = json.NewEncoder(indexFile).Encode(index)
	if err != nil {
		indexFile.Close()
		return fmt.Errorf("encode index file: %w", err)
	}

	err = indexFile.Close()
	if err != nil {
		return fmt.Errorf("close index file: %w", err)
	}
	return nil
}
