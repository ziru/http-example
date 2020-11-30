package main

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

var (
	ErrNotFound = errors.New("not found")

	defaultTTL              = time.Second * 10
	defaultEvictionInterval = time.Second * 2
)

type entry struct {
	val        interface{}
	expiration time.Time
}

type cache struct {
	ctx   context.Context
	ttl   time.Duration
	items sync.Map
}

func NewCache(ctx context.Context) *cache {
	c := &cache{
		ctx: ctx,
		ttl: defaultTTL,
	}
	go c.startEvicter(defaultEvictionInterval)
	return c
}

func (c *cache) Put(key interface{}, val interface{}) error {
	c.items.Store(key, &entry{
		val:        val,
		expiration: time.Now().Add(c.ttl),
	})
	return nil
}

func (c *cache) Get(key interface{}) (interface{}, error) {
	item, ok := c.items.Load(key)
	if !ok {
		return nil, ErrNotFound
	}
	e := item.(*entry)
	return e.val, nil
}

func (c *cache) startEvicter(evictionInterval time.Duration) {
	ticker := time.NewTicker(evictionInterval)
	defer ticker.Stop()
	for {
		select {
		case now := <-ticker.C:
			c.evictExpired(now)
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *cache) evictExpired(now time.Time) {
	// log.Printf("Attempt to evict expired entries")
	c.items.Range(func(key, value interface{}) bool {
		// log.Printf("> Check key:%v", key)
		e := value.(*entry)
		if e.expiration.Before(now) {
			log.Printf(">> Expired: key=%v", key)
			c.items.Delete(key)
		}
		return true
	})
}
