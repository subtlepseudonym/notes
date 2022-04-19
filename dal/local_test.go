package dal

import (
	"fmt"
	"strings"
	"os"
	"io"
	"testing"
	"time"
	"path"

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

func TestLocalCreateNote(t *testing.T) {
	tmpDir := t.TempDir()
	noteDir := path.Join(tmpDir, "notes")

	localDAL, err := NewLocal(noteDir)
	if err != nil {
		t.Error(err)
	}

	now := time.Now()

	id := "82ed62cf-fa24-43cd-8468-1877ef20858f"
	title := "Test Note"
	body := "This is a test note body"
	createdAt := now
	updatedAt := now.Add(time.Minute)
	tags := []string{"test", "tags", "медведь"}

	note := notes.Note{
		ID: id,
		Title: title,
		Body: body,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Tags: tags,
	}

	err = localDAL.CreateNote(&note)
	if err != nil {
		t.Error(err)
	}

	notePath := path.Join(noteDir, id)
	noteFile, err := os.Open(notePath)
	if err != nil {
		t.Error(err)
	}
	defer noteFile.Close()

	readBody, err := io.ReadAll(noteFile)
	if err != nil {
		t.Error(err)
	}

	if string(readBody) != body {
		t.Errorf("unexpected note body: %q != %q", string(readBody), body)
	}

	metaPath := path.Join(noteDir, fmt.Sprintf("%s.meta", id))
	metaFile, err := os.Open(metaPath)
	if err != nil {
		t.Error(err)
	}
	defer metaFile.Close()

	var readMeta meta
	_, err = toml.NewDecoder(metaFile).Decode(&readMeta)
	if err != nil {
		t.Error(err)
	}

	if readMeta.ID != id {
		t.Errorf("unexpected id: %q != %q", readMeta.ID, id)
	}
	if readMeta.Title != title {
		t.Errorf("unexpected title: %q != %q", readMeta.Title, title)
	}
	if !readMeta.CreatedAt.Equal(createdAt) {
		t.Errorf("unexpected createdAt: %q != %q", readMeta.CreatedAt, createdAt)
	}
	if !readMeta.UpdatedAt.Equal(updatedAt) {
		t.Errorf("unexpected updatedAt: %q != %q", readMeta.UpdatedAt, updatedAt)
	}

	readMetaTags := strings.Join(readMeta.Tags, ",")
	expectedTags := strings.Join(tags, ",")
	if readMetaTags != expectedTags {
		t.Errorf("unexpected tags: %q != %q", readMetaTags, expectedTags)
	}
}

func TestLocalReadNote(t *testing.T) {
}

func TestLocalUpdateNote(t *testing.T) {
}

func TestLocalDeleteNote(t *testing.T) {
}
