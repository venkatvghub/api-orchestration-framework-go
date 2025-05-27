package utils

import (
	"bytes"
	"sync"
)

// BufferPool provides a pool of reusable byte buffers
type BufferPool struct {
	pool sync.Pool
}

// NewBufferPool creates a new buffer pool
func NewBufferPool() *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

// Get retrieves a buffer from the pool
func (p *BufferPool) Get() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer)
}

// Put returns a buffer to the pool after resetting it
func (p *BufferPool) Put(buf *bytes.Buffer) {
	buf.Reset()
	p.pool.Put(buf)
}

// MapPool provides a pool of reusable maps
type MapPool struct {
	pool sync.Pool
}

// NewMapPool creates a new map pool
func NewMapPool() *MapPool {
	return &MapPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make(map[string]interface{})
			},
		},
	}
}

// Get retrieves a map from the pool
func (p *MapPool) Get() map[string]interface{} {
	return p.pool.Get().(map[string]interface{})
}

// Put returns a map to the pool after clearing it
func (p *MapPool) Put(m map[string]interface{}) {
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	p.pool.Put(m)
}

// StringSlicePool provides a pool of reusable string slices
type StringSlicePool struct {
	pool sync.Pool
}

// NewStringSlicePool creates a new string slice pool
func NewStringSlicePool() *StringSlicePool {
	return &StringSlicePool{
		pool: sync.Pool{
			New: func() interface{} {
				s := make([]string, 0, 10) // Pre-allocate capacity
				return &s
			},
		},
	}
}

// Get retrieves a string slice from the pool
func (p *StringSlicePool) Get() []string {
	return *p.pool.Get().(*[]string)
}

// Put returns a string slice to the pool after resetting it
func (p *StringSlicePool) Put(s []string) {
	s = s[:0] // Reset length but keep capacity
	p.pool.Put(&s)
}

// Global pools for common use cases
var (
	GlobalBufferPool      = NewBufferPool()
	GlobalMapPool         = NewMapPool()
	GlobalStringSlicePool = NewStringSlicePool()
)

// Convenience functions for global pools
func GetBuffer() *bytes.Buffer {
	return GlobalBufferPool.Get()
}

func PutBuffer(buf *bytes.Buffer) {
	GlobalBufferPool.Put(buf)
}

func GetMap() map[string]interface{} {
	return GlobalMapPool.Get()
}

func PutMap(m map[string]interface{}) {
	GlobalMapPool.Put(m)
}

func GetStringSlice() []string {
	return GlobalStringSlicePool.Get()
}

func PutStringSlice(s []string) {
	GlobalStringSlicePool.Put(s)
}
