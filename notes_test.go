package notes

import (
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
		body, err := GetNoteBodyFromUser(test.EditorCommand, test.ExistingBody)
		if err != nil {
			t.Error(err)
		}

		if body != test.ExpectedBody {
			t.Errorf("body %q != expected body %q", body, test.ExpectedBody)
		}
	}
}
