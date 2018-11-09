package files

import (
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

const defaultNotesDir = ".notes"

// GetDefaultNotesDir returns the path to the default directory containing notes files
func GetDefaultNotesDir() (string, error) {
	if dir := os.Getenv(notesDirEnvVar); dir != "" {
		return dir, nil
	}

	home, err := homedir.Dir()
	if err != nil {
		return "", errors.Wrap(err, "get home directory failed")
	}

	return home + defaultNotesDir
}
