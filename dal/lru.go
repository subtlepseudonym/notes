package dal

import (
)

type lru struct {
	capacity int
	front *node
	rear *node
}

func newLRU(capacity int) cache {
	return lru{
		capacity: capacity,
		// node init
	}
}

func (l lru) Get(idx int) *notes.Note {
}

func (l lru) enqueue(note *notes.Note) {
}

func (l lru) dequeue() {
}
