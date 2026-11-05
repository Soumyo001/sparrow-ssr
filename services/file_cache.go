package services

import "sync"

// fileCache keeps one immutable decoded value per relative storage path.
// Concurrent requests for a missing file share a single load operation.
type fileCache[T any] struct {
	mu      sync.Mutex
	entries map[string]*fileEntry[T]
}

type fileEntry[T any] struct {
	ready chan struct{}
	value T
	err   error
}

func newFileCache[T any]() *fileCache[T] {
	return &fileCache[T]{
		entries: make(map[string]*fileEntry[T]),
	}
}

func (c *fileCache[T]) getOrLoad(
	key string,
	fetch func() (T, error),
) (T, error) {
	c.mu.Lock()

	if existing, found := c.entries[key]; found {
		c.mu.Unlock()

		// Another request may already be loading this file.
		// Wait for that operation instead of reading again.
		<-existing.ready

		return existing.value, existing.err
	}

	// The first request owns the load operation.
	entry := &fileEntry[T]{
		ready: make(chan struct{}),
	}

	c.entries[key] = entry
	c.mu.Unlock()

	// Perform I/O without holding the cache mutex.
	value, err := fetch()

	c.mu.Lock()

	entry.value = value
	entry.err = err

	// Failed reads should be retried on future requests.
	if err != nil && c.entries[key] == entry {
		delete(c.entries, key)
	}

	// Publish the result to waiting goroutines.
	close(entry.ready)

	c.mu.Unlock()

	return value, err
}

// getReady never blocks.
// It allows a fast path when a file is already cached.
func (c *fileCache[T]) getReady(key string) (T, bool) {
	var zero T

	c.mu.Lock()
	entry, found := c.entries[key]
	c.mu.Unlock()

	if !found {
		return zero, false
	}

	select {
	case <-entry.ready:
		if entry.err != nil {
			return zero, false
		}

		return entry.value, true

	default:
		// The file is still loading.
		return zero, false
	}
}

func (c *fileCache[T]) invalidate(key string) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}

func (c *fileCache[T]) size() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.entries)
}
