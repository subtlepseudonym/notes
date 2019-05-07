package notes

import (
	"testing"

	"bou.ke/monkey"
)

func TestNoteAppendEdit(t *testing.T) {
}

type NoteBodyTest struct {
	EditorCommand string
	ExistingBody  string

	BodyToWrite string

	ExpectedBody  string
	ErrorExpected bool
}

func TestGetNoteBodyFromUser(t *testing.T) {
	tests := []NoteBodyTest{
		// TODO: add tests for failure cases
		// TODO: monkey patch os/exec
		{
			EditorCommand: "editor",
			BodyToWrite:   "write me",
			ExpectedBody:  "write me",
		},
	}

	for _, test := range tests {
		body, err := GetNoteBodyFromUser(test.EditorCommand, test.ExistingBody)
		if test.ErrorExpected {
			if err != nil {
				continue
			}
			t.Error("expected non-nil error, got nil")
		} else if err != nil {
			t.Error(err)
		}

		if body != test.ExpectedBody {
			t.Errorf("body %q != expected body %q", body, test.ExpectedBody)
		}
	}
}
