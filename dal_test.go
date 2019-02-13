package notes

import (
	"path"
	"testing"

	"github.com/go-test/deep"
	"github.com/mitchellh/go-homedir"
)

func TestNewDefaultDAL(t *testing.T) {
	version := "totally not a semantic version"
	dal, err := NewDefaultDAL(version)
	if err != nil {
		t.Error(err)
	}

	home, err := homedir.Dir()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	expected := &defaultDAL{
		version:            version,
		metaFilename:       defaultMetaFilename,
		notesDirectoryPath: path.Join(home, defaultNotesDirectory),
		noteFilenameFormat: defaultNoteFilenameFormat,
	}

	if diff := deep.Equal(dal, expected); diff != nil {
		t.Error(diff)
	}
}

func TestDefaultDALGetMeta(t *testing.T) {
}

func TestDefaultDALSaveMeta(t *testing.T) {
}

func TestDefaultDALGetNote(t *testing.T) {
}

func TestDefaultDALSaveNote(t *testing.T) {
}

func TestDefaultDALRemoveNote(t *testing.T) {
}
