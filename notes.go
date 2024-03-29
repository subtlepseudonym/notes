package notes

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	defaultEditHistorySize = 16
)

// Meta holds meta information for the local notes storage as a whole
type Meta struct {
	Version     string   `json:"version"`
	OldVersions []string `json:"oldVersions"`
	LatestID    int      `json:"latestID"`
	Size        int      `json:"size"` // meta file size in bytes
}

// UpdateVersion replaces the existing version with the provided new version
// and adds the old version to the list of old versions
func (m *Meta) UpdateVersion(newVersion string) *Meta {
	m.OldVersions = append(m.OldVersions, m.Version)
	m.Version = newVersion

	return m
}

// ApproxSize gets the approximate encoded size of the meta object by
// encoding it and returning its byte length
// NOTE: This will be off by the byte length of one numeric character as
// compared to its size on disk when used via the nts cli
func (m Meta) ApproxSize() (int, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return 0, fmt.Errorf("encode meta: %w", err)
	}

	return len(b), nil
}

type JSONTime struct {
	time.Time
}

func (j JSONTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.UnixNano()) // FIXME: this will fail after 2262
}

func (j *JSONTime) UnmarshalJSON(b []byte) error {
	var i int64
	err := json.Unmarshal(b, &i)
	if err != nil {
		return err
	}

	j.Time = time.Unix(0, i)
	return nil
}

// NoteMeta holds meta information for one note to make commands that only access
// meta information perform faster
type NoteMeta struct {
	ID      int           `json:"id"`
	Title   string        `json:"title"`
	Created JSONTime      `json:"created"`
	Deleted JSONTime      `json:"deleted"`
	History []EditHistory `json:"history"`
}

// EditHistory holds meta information that changes over time
type EditHistory struct {
	Updated JSONTime `json:"updated"`
	Size    int      `json:"size"` // file size in bytes
}

// Note includes the content of the note as well as its meta information as backup in
// case we need to recreate the meta file from scratch
type Note struct {
	Meta NoteMeta `json:"meta"`
	Body string   `json:"body"`
}

// ApproxSize gets the approximate encoded size of the note object by encoding
// it and returning its byte length
// NOTE: This is likely to be a bit lower than the actual size on disk because it is
// intended for use with the nts cli, which alters the note's edit history before
// encoding and saving it to disk
func (n Note) ApproxSize() (int, error) {
	b, err := json.Marshal(n)
	if err != nil {
		return 0, fmt.Errorf("encode note: %w", err)
	}

	return len(b), nil
}

// AppendEdit adds a new EditHistory to the note's history, trimming if the history
// exceeds the edit history size limit
func (n *Note) AppendEdit(timestamp time.Time) (*Note, error) {
	noteSize, err := n.ApproxSize()
	if err != nil {
		return n, fmt.Errorf("get note size: %w", err)
	}

	update := EditHistory{
		Updated: JSONTime{time.Now()},
		Size:    noteSize,
	}
	n.Meta.History = append([]EditHistory{update}, n.Meta.History...)
	if len(n.Meta.History) > defaultEditHistorySize {
		n.Meta.History = n.Meta.History[:defaultEditHistorySize]
	}

	return n, nil
}
