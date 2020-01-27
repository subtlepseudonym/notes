package dal

import (
	"path"
	"testing"

	"github.com/go-test/deep"
	"github.com/mitchellh/go-homedir"
)

func TestNewLocalDAL(t *testing.T) {
	version := "totally not a semantic version"
	dir := "notes_test_dir"
	dal, err := NewLocalDAL(dir, version)
	if err != nil {
		t.Error(err)
	}

	home, err := homedir.Dir()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	expected := &localDAL{
		version:            version,
		metaFilename:       defaultMetaFilename,
		notesDirectoryPath: path.Join(home, dir),
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
