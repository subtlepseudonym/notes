package dal

import (
	"github.com/subtlepseudony/notes"
)

type cache interface {
	Get(int) *notes.Note
}

type CacheType int
const (
	LRU = iota
)

type dalCache struct {
	DAL
	meta *notes.Meta
	recentlyUsed cache
	recentlyCreated cache
}

func WithCache(dal DAL, cacheType CacheType, recentlyUsedCap, recentlyCreatedCap int) DAL {
	var usedCache cache
	var createdCache cache

	switch cacheType {
	case LRU:
		usedCache = newLRU(recentlyUsedCap)
		createdCache = newLRU(recentlyCreatedCap)
	default: // defaults to LRU
		usedCache = newLRU(recentlyUsedCap)
		createdCache = newLRU(recentlyCreatedCap)
	}

	return &dalCache{
		DAL: dal,
		recentlyUsed: usedCache,
		recentlyCreated: createdCache,
	}
}

func (c *dalCache) GetMeta() (*notes.Meta, error) {
	if c.meta != nil {
		return c.meta, nil
	}

	meta, err := c.dal.GetMeta()
	if err == nil {
		c.meta = meta
	}

	return meta, err
}

func (c *dalCache) SaveMeta(meta *notes.Meta) error {
	err := c.dal.SaveMeta(meta)
	if err == nil {
		c.meta = meta
	}

	return err
}

func (c *dalCache) GetNote(id int) (*notes.Note, error) {
	note, err := c.dal.GetNote(id)
	if err == nil {
		c.notes.enqueue(note)
	}

	return note, err
}

func (c *dalCache) SaveNote(note *notes.Note) error {
	err := c.dal.SaveNote(note)
	if err == nil {
		c.notes.enqueue(note)
	}

	return err
}

func (c *dalCache) RemoveNote(id int) error {
	return c.dal.RemoveNote(id)
}
