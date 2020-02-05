package cache

import (
	"github.com/subtlepseudonym/notes/dal"
)

type CacheType int

const (
	Noop CacheType = iota
	LRU            // least recently used
	RR             // random replacement
)

type NoteCache interface {
	dal.DAL
	Flush() error
}

func NewNoteCache(d dal.DAL, cacheType CacheType, capacity int) NoteCache {
	switch cacheType {
	case LRU:
		return NewLeastRecentlyUsed(d, capacity)
	case RR:
		return NewRandomReplacement(d, capacity)
	default:
		return NewNoop(d)
	}
}
