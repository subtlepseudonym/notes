package notes

import (
	"time"
)

type Note struct {
	ID string
	Title string
	Body string
	CreatedAt time.Time
	UpdatedAt time.Time
	Tags []string
}
