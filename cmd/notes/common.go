package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/subtlepseudonym/notes"

	"github.com/urfave/cli"
	"go.uber.org/zap"
)

const (
	defaultEditor       = "vim"
	defaultUpdatePeriod = 5 * time.Minute
)

// editNote is a helper function for turning control over to the user and getting
// a new note body from them
func (a *App) editNote(ctx *cli.Context, note *notes.Note, logger *zap.Logger) (string, error) {
	file, err := ioutil.TempFile("", "note")
	if err != nil {
		return "", fmt.Errorf("create temporary file: %w", err)
	}
	defer file.Close()

	stop := make(chan struct{})
	if !ctx.Bool("no-watch") {
		go func() {
			err := a.watchAndUpdate(ctx, note, file.Name(), ctx.Duration("update-period"), stop, logger)
			if err != nil {
				a.logger.Error("watch and update failed", zap.Error(err), zap.Int("noteID", note.Meta.ID), zap.String("filename", file.Name()))
			}
		}()
	}

	body, err := getNoteBodyFromUser(file, ctx.String("editor"), note.Body)
	if err != nil {
		return "", fmt.Errorf("get note body from user: %w", err)
	}
	close(stop)

	return body, nil
}

// watchAndUpdate periodically reads the contents of the provided file and compares
// it to the body of the provided note. If they aren't equal, it saves the changes
// to the DAL
func (a *App) watchAndUpdate(ctx *cli.Context, note *notes.Note, filename string, period time.Duration, stop chan struct{}, l *zap.Logger) error {
	logger := l.Named("watch")

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return nil
		case timestamp := <-ticker.C:
			b, err := ioutil.ReadFile(filename)
			if err != nil {
				return fmt.Errorf("read file: %w", err) // FIXME: might want to log these rather than returning
			}

			if bytes.Equal(b, []byte(note.Body)) {
				continue
			}

			note.Body = string(b)
			note, err = note.AppendEdit(timestamp)
			if err != nil {
				return fmt.Errorf("append edit to history: %w", err)
			}

			err = a.dal.SaveNote(note)
			if err != nil {
				return fmt.Errorf("save note: %w", err)
			}
			logger.Info("note updated", zap.Int("noteID", note.Meta.ID))

			a.index[note.Meta.ID] = note.Meta
			err = a.dal.SaveIndex(a.index)
			if err != nil {
				return fmt.Errorf("save index: %w", err)
			}
		}
	}

	return nil
}

// getNoteBodyFromUser drops the user into the provided editor command before
// retrieving the contents of the edited file
func getNoteBodyFromUser(file *os.File, editor, existingBody string) (string, error) {
	_, err := fmt.Fprint(file, existingBody)
	if err != nil {
		return "", fmt.Errorf("print existing body to temporary file: %w", err)
	}

	cmd := exec.Command(editor, file.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("run editor command: %w", err)
	}

	bodyBytes, err := ioutil.ReadFile(file.Name())
	if err != nil {
		return "", fmt.Errorf("read temporary file: %w", err)
	}

	return string(bodyBytes), nil
}
