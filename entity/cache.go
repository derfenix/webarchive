package entity

import (
	"bytes"
	"io"
	"sync"
)

func NewCache() *Cache {
	return &Cache{data: make([]byte, 0, 1024*512)}
}

type Cache struct {
	mu   sync.RWMutex
	data []byte
}

func (c *Cache) Write(p []byte) (n int, err error) {
	c.mu.Lock()
	c.data = append(c.data, p...)
	c.mu.Unlock()

	return len(p), nil
}

func (c *Cache) Get() []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.data
}

func (c *Cache) Reader() io.Reader {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.data) == 0 {
		return nil
	}

	return bytes.NewBuffer(c.data)
}
