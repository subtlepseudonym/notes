package cache

import (
	"fmt"
	"math/rand"

	"github.com/subtlepseudonym/notes"
	"github.com/subtlepseudonym/notes/dal"
)

// rr uses a random replacement cache replacement policy
type rr struct {
	dal.DAL
	capacity int
	index []int
	cache map[int]*notes.Note
}

func NewRR(d dal.DAL, capacity int) rr {
	return rr{
		DAL: d,
		capacity: capacity,
		index: make([]int, 0, capacity),
		cache: make(map[int]*notes.Note, capacity),
	}
}

func (r rr) Flush() error {
	r.index = make([]int, 0, r.capacity)
	r.cache = make(map[int]*notes.Note, r.capacity)

	return nil
}

func (r rr) GetNote(id int) (*notes.Note, error) {
	cached, exists := r.cache[id]
	if exists {
		return cached, nil
	}

	note, err := r.DAL.GetNote(id)
	if err != nil {
		return nil, fmt.Errorf("cache miss: %w", err)
	}

	r.add(note)
	return note, nil
}

func (r rr) add(note *notes.Note) {
	if len(r.index) < r.capacity {
		r.index = append(r.index, note.Meta.ID)
	} else {
		idx := rand.Intn(len(r.index))
		noteID := r.index[idx]
		delete(r.cache, noteID)

		r.index[idx] = note.Meta.ID
	}

	r.cache[note.Meta.ID] = note
}
