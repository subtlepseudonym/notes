package dal

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"sync"
	"time"

	"github.com/subtlepseudonym/notes"

	"gopkg.in/yaml.v3"
)

// meta is an internal representation of non-body notes.Note fields
type meta struct {
	ID        string    `yaml:"id"`
	Title     string    `yaml:"title"`
	CreatedAt time.Time `yaml:"created_at"`
	UpdatedAt time.Time `yaml:"updated_at"`
	Tags      []string  `yaml:"tags"`
}

func noteToMeta(note *notes.Note) meta {
	return meta{
		ID:        note.ID,
		Title:     note.Title,
		CreatedAt: note.CreatedAt,
		UpdatedAt: note.UpdatedAt,
		Tags:      note.Tags,
	}
}

func metaToNote(m meta) notes.Note {
	return notes.Note{
		ID:        m.ID,
		Title:     m.Title,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		Tags:      m.Tags,
	}
}

type local struct {
	sync.RWMutex
	root string
}

func NewLocal(directory string) (DAL, error) {
	err := os.Mkdir(directory, os.ModeDir|os.FileMode(0700))
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return nil, fmt.Errorf("make directory: %w", err)
	}

	return &local{
		root: directory,
	}, nil
}

func (d *local) ReadNote(id string) (*notes.Note, error) {
	d.RLock()
	defer d.RUnlock()

	notePath := path.Join(d.root, id)
	noteFile, err := os.Open(notePath)
	if err != nil {
		return nil, fmt.Errorf("open note file: %w", err)
	}
	defer noteFile.Close()

	b, err := io.ReadAll(noteFile)
	if err != nil {
		return nil, fmt.Errorf("read note file: %w", err)
	}
	body := string(b)

	metaPath := path.Join(d.root, fmt.Sprintf("%s.meta", id))
	metaFile, err := os.Open(metaPath)
	if err != nil {
		return nil, fmt.Errorf("open meta file: %w", err)
	}
	defer metaFile.Close()

	var noteMeta meta
	_, err = yaml.NewDecoder(metaFile).Decode(&noteMeta)
	if err != nil {
		return nil, fmt.Errorf("decode meta file: %w", err)
	}

	note := metaToNote(noteMeta)
	note.Body = body

	return &note, nil
}

func (d *local) WriteNote(note *notes.Note) error {
	d.Lock()
	defer d.Unlock()

	notePath := path.Join(d.root, note.ID)
	err := os.Rename(notePath, fmt.Sprintf("%s.bak", notePath))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("back up note file: %w", err)
	}

	metaPath := path.Join(d.root, fmt.Sprintf("%s.meta", note.ID))
	err = os.Rename(metaPath, fmt.Sprintf("%s.bak", metaPath))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("back up meta file: %w", err)
	}

	noteFile, err := os.Create(notePath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer noteFile.Close()

	_, err = noteFile.WriteString(note.Body)
	if err != nil {
		return fmt.Errorf("write body: %w", err)
	}

	metaFile, err := os.Create(metaPath)
	if err != nil {
		return fmt.Errorf("create meta file: %w", err)
	}
	defer metaFile.Close()

	noteMeta := noteToMeta(note)
	err = yaml.NewEncoder(metaFile).Encode(noteMeta)
	if err != nil {
		return fmt.Errorf("write meta: %w", err)
	}

	return nil
}

func (d *local) DeleteNote(id string, hard bool) error {
	d.Lock()
	defer d.Unlock()

	notePath := path.Join(d.root, id)
	err := os.Remove(notePath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove note file: %w", err)
	}

	if hard {
		err = os.Remove(fmt.Sprintf("%s.bak", notePath))
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("remove backup note file: %w", err)
		}
	}

	metaPath := path.Join(d.root, fmt.Sprintf("%s.meta", id))
	err = os.Remove(metaPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove meta file: %w", err)
	}

	if hard {
		err = os.Remove(fmt.Sprintf("%s.bak", metaPath))
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("remove backup meta file: %w", err)
		}
	}

	return nil
}
