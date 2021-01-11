package dal

import (
	"os"
	"path"
	"testing"

	"github.com/go-test/deep"
)

func TestNewLocalDAL(t *testing.T) {
	version := "totally not a semantic version"
	dir := "notes_test_dir"
	dal, err := NewLocal(dir, version)
	if err != nil {
		t.Error(err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	expected := &local{
		baseDirectory:      path.Join(home, dir),
		notebook:           defaultNotebook,
		metaFilename:       defaultMetaFilename,
		noteFilenameFormat: defaultNoteFilenameFormat,
	}

	if diff := deep.Equal(dal, expected); diff != nil {
		t.Error(diff)
	}
}

func TestLocalDALGetMeta(t *testing.T) {
}

func TestLocalDALSaveMeta(t *testing.T) {
}

func TestLocalDALGetNote(t *testing.T) {
}

func TestLocalDALSaveNote(t *testing.T) {
}

func TestLocalDALRemoveNote(t *testing.T) {
}
