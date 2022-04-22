package dal

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/subtlepseudonym/notes"

	"github.com/BurntSushi/toml"
)

func TestNewLocal(t *testing.T) {
	tmpDir := t.TempDir()
	noteDir := path.Join(tmpDir, "notes")

	dal, err := NewLocal(noteDir)
	if err != nil {
		t.Error(err)
	}

	localDAL, _ := dal.(*local)

	if localDAL.root != noteDir {
		t.Errorf("unexpected localDAL.root (%q != %q)", localDAL.root, noteDir)
	}

	info, err := os.Stat(noteDir)
	if err != nil {
		t.Error(err)
	}

	if !info.IsDir() {
		t.Errorf("file %q is not a directory", noteDir)
	}
}

func TestLocalReadNote(t *testing.T) {
	tmpDir := t.TempDir()
	noteDir := path.Join(tmpDir, "notes")

	localDAL, err := NewLocal(noteDir)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	localDAL, err = NewLocal(noteDir)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	now := time.Now()

	id := "00QBTG0FERFTCCNFISUBO489DI"
	title := "Test note"
	body := "Testing notes with text"
	createdAt := now
	updatedAt := now.Add(time.Minute)
	tags := []string{"lots", "of", "tags"}

	notePath := path.Join(noteDir, id)
	noteFile, err := os.Create(notePath)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	_, err = noteFile.WriteString(body)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	metaPath := path.Join(noteDir, fmt.Sprintf("%s.meta", id))
	metaFile, err := os.Create(metaPath)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	m := meta{
		ID:        id,
		Title:     title,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Tags:      tags,
	}

	err = toml.NewEncoder(metaFile).Encode(&m)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	note, err := localDAL.ReadNote(id)
	if err != nil {
		t.Error()
		t.FailNow()
	}

	if note.ID != id {
		t.Errorf("unexpected id: %q != %q", note.ID, id)
	}
	if note.Title != title {
		t.Errorf("unexpected title: %q != %q", note.Title, title)
	}
	if note.Body != body {
		t.Errorf("unexpected body: %q != %q", note.Body, body)
	}
	if !note.CreatedAt.Equal(createdAt) {
		t.Errorf("unexpected createdAt: %s != %s", note.CreatedAt, createdAt)
	}
	if !note.UpdatedAt.Equal(updatedAt) {
		t.Errorf("unexpected updatedAt: %s != %s", note.UpdatedAt, updatedAt)
	}

	expectedTags := strings.Join(tags, ",")
	actualTags := strings.Join(note.Tags, ",")
	if actualTags != expectedTags {
		t.Errorf("unexpected tags: %q != %q", actualTags, expectedTags)
	}
}

func TestLocalWriteNote(t *testing.T) {
	tmpDir := t.TempDir()
	noteDir := path.Join(tmpDir, "notes")

	localDAL, err := NewLocal(noteDir)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	now := time.Now()

	id := "82ed62cf-fa24-43cd-8468-1877ef20858f"
	title := "Test Note"
	body := "This is a test note body"
	createdAt := now
	updatedAt := now.Add(time.Minute)
	tags := []string{"test", "tags", "медведь"}

	note := notes.Note{
		ID:        id,
		Title:     title,
		Body:      body,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Tags:      tags,
	}

	err = localDAL.WriteNote(&note)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	notePath := path.Join(noteDir, id)
	noteFile, err := os.Open(notePath)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer noteFile.Close()

	readBody, err := io.ReadAll(noteFile)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if string(readBody) != body {
		t.Errorf("unexpected note body: %q != %q", string(readBody), body)
		t.FailNow()
	}

	metaPath := path.Join(noteDir, fmt.Sprintf("%s.meta", id))
	metaFile, err := os.Open(metaPath)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer metaFile.Close()

	var readMeta meta
	_, err = toml.NewDecoder(metaFile).Decode(&readMeta)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if readMeta.ID != id {
		t.Errorf("unexpected id: %q != %q", readMeta.ID, id)
	}
	if readMeta.Title != title {
		t.Errorf("unexpected title: %q != %q", readMeta.Title, title)
	}
	if !readMeta.CreatedAt.Equal(createdAt) {
		t.Errorf("unexpected createdAt: %s != %s", readMeta.CreatedAt, createdAt)
	}
	if !readMeta.UpdatedAt.Equal(updatedAt) {
		t.Errorf("unexpected updatedAt: %s != %s", readMeta.UpdatedAt, updatedAt)
	}

	readMetaTags := strings.Join(readMeta.Tags, ",")
	expectedTags := strings.Join(tags, ",")
	if readMetaTags != expectedTags {
		t.Errorf("unexpected tags: %q != %q", readMetaTags, expectedTags)
	}
}

func TestLocalDeleteNote(t *testing.T) {
}
