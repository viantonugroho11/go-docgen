package pdf

import (
	"container/list"
	"crypto/sha256"
	"sync"
)

// pdfCache is a fixed-capacity LRU keyed by sha256(html). Values are
// pre-rendered PDF byte slices. All exported methods are safe for concurrent
// use. A capacity of 0 disables the cache (all lookups miss, stores are no-ops).
type pdfCache struct {
	mu    sync.Mutex
	cap   int
	items map[[32]byte]*list.Element
	order *list.List // front = most recently used
}

type cacheEntry struct {
	key [32]byte
	pdf []byte
}

func newCache(capacity int) *pdfCache {
	if capacity <= 0 {
		return &pdfCache{}
	}
	return &pdfCache{
		cap:   capacity,
		items: make(map[[32]byte]*list.Element, capacity),
		order: list.New(),
	}
}

func (c *pdfCache) enabled() bool { return c.cap > 0 }

func (c *pdfCache) get(html string) ([]byte, bool) {
	if !c.enabled() {
		return nil, false
	}
	k := sha256.Sum256([]byte(html))
	c.mu.Lock()
	defer c.mu.Unlock()
	el, ok := c.items[k]
	if !ok {
		return nil, false
	}
	c.order.MoveToFront(el)
	return el.Value.(*cacheEntry).pdf, true
}

// put stores a copy-on-return snapshot of pdf keyed by html. The caller may
// mutate the returned pdf slice safely; the cache retains its own copy.
func (c *pdfCache) put(html string, pdf []byte) {
	if !c.enabled() || len(pdf) == 0 {
		return
	}
	k := sha256.Sum256([]byte(html))
	stored := make([]byte, len(pdf))
	copy(stored, pdf)
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[k]; ok {
		el.Value.(*cacheEntry).pdf = stored
		c.order.MoveToFront(el)
		return
	}
	el := c.order.PushFront(&cacheEntry{key: k, pdf: stored})
	c.items[k] = el
	if c.order.Len() > c.cap {
		oldest := c.order.Back()
		if oldest != nil {
			c.order.Remove(oldest)
			delete(c.items, oldest.Value.(*cacheEntry).key)
		}
	}
}
