package notes

import (
	"io/ioutil"
	"testing"
)

type NoteBodyTest struct {
	EditorCommand string
	ExistingBody  string

	ExpectedBody string
}

func TestGetNoteBodyFromUser(t *testing.T) {
	tests := []NoteBodyTest{
		// TODO: add tests for failure cases
		// TODO: monkey patch os/exec
	}

	for _, test := range tests {
		editFile, err := ioutil.TempFile("", "test")
		if err != nil {
			t.Error(err)
		}

		body, err := GetNoteBodyFromUser(editFile, test.EditorCommand, test.ExistingBody)
		if err != nil {
			t.Error(err)
		}

		if body != test.ExpectedBody {
			t.Errorf("body %q != expected body %q", body, test.ExpectedBody)
		}
	}
}
