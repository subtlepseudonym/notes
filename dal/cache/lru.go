package cache

import (
	"fmt"

	"github.com/subtlepseudonym/notes"
	"github.com/subtlepseudonym/notes/dal"
)

// lru utilizes a least-recently-used cache replacement policy
type lru struct {
	dal.DAL
	capacity int
	index    map[int]*node // map noteID to linked list pointer
	front    *node
	rear     *node
}

type node struct {
	prev *node
	next *node
	note *notes.Note
}

func NewLeastRecentlyUsed(d dal.DAL, capacity int) NoteCache {
	return NewLRU(d, capacity)
}

func NewLRU(d dal.DAL, capacity int) NoteCache {
	return lru{
		DAL:      d,
		capacity: capacity,
		index:    make(map[int]*node, capacity),
	}
}

func (l lru) Flush() error {
	l.index = make(map[int]*node, l.capacity)
	l.front = nil
	l.rear = nil

	return nil
}

func (l lru) GetNote(id int) (*notes.Note, error) {
	cached, exists := l.index[id]
	if exists {
		l.moveToFront(cached)
		return cached.note, nil
	}

	note, err := l.DAL.GetNote(id)
	if err != nil {
		return nil, fmt.Errorf("cache miss: %w", err)
	}

	newNode := node{
		note: note,
	}
	l.add(&newNode)

	return note, nil
}

func (l lru) moveToFront(n *node) {
	if n == l.front {
		return
	}

	n.prev.next = n.next
	if n.next != nil {
		n.next.prev = n.prev
	}

	if n == l.rear {
		l.rear = n.prev
		l.rear.next = nil
	}

	n.next = l.front
	n.prev = nil

	n.next.prev = n
	l.front = n
}

func (l lru) add(n *node) {
	if l.front == nil {
		l.front = n
		l.rear = n

		return
	}

	n.next = l.front
	l.front.prev = n

	l.front = n
	l.index[n.note.Meta.ID] = n

	if len(l.index)+1 > l.capacity {
		l.removeOldest()
	}
}

func (l lru) removeOldest() {
	if len(l.index) == 0 {
		return
	}

	if l.front == l.rear {
		l.front = nil
		l.rear = nil
		return
	}

	id := l.rear.note.Meta.ID
	delete(l.index, id)

	l.rear = l.rear.prev
	if l.rear != nil {
		l.rear.next = nil
	}
}
