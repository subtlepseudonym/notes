package cache

import (
	"github.com/subtlepseudonym/notes/dal"
)

type CacheType int
const (
	Noop CacheType = iota
	LRU // least recently used
	RR  // random replacement
)

type NoteCache interface {
	dal.DAL
	Flush() error
}

func WithNoteCache(d dal.DAL, cacheType CacheType, capacity int) NoteCache {
	switch cacheType {
	case LRU:
		return NewLRU(d, capacity)
	case RR:
		return NewRR(d, capacity)
	default:
		return NewNoop(d)
	}
}
