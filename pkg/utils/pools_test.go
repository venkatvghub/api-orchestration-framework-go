package utils

import (
	"bytes"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBufferPool(t *testing.T) {
	t.Run("NewBufferPool", func(t *testing.T) {
		pool := NewBufferPool()
		assert.NotNil(t, pool)
	})

	t.Run("Get and Put", func(t *testing.T) {
		pool := NewBufferPool()

		// Get a buffer
		buf := pool.Get()
		assert.NotNil(t, buf)
		assert.IsType(t, &bytes.Buffer{}, buf)

		// Write to buffer
		buf.WriteString("test data")
		assert.Equal(t, "test data", buf.String())

		// Put buffer back
		pool.Put(buf)

		// Buffer should be reset
		assert.Equal(t, 0, buf.Len())
		assert.Equal(t, "", buf.String())
	})

	t.Run("Multiple buffers", func(t *testing.T) {
		pool := NewBufferPool()

		buf1 := pool.Get()
		buf2 := pool.Get()

		buf1.WriteString("buffer1")
		buf2.WriteString("buffer2")

		assert.Equal(t, "buffer1", buf1.String())
		assert.Equal(t, "buffer2", buf2.String())

		pool.Put(buf1)
		pool.Put(buf2)

		// Get buffers again - they should be clean
		buf3 := pool.Get()
		buf4 := pool.Get()

		assert.Equal(t, "", buf3.String())
		assert.Equal(t, "", buf4.String())
	})

	t.Run("Concurrent access", func(t *testing.T) {
		pool := NewBufferPool()
		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				buf := pool.Get()
				buf.WriteString("test")
				assert.Equal(t, "test", buf.String())
				pool.Put(buf)
			}(i)
		}

		wg.Wait()
	})
}

func TestMapPool(t *testing.T) {
	t.Run("NewMapPool", func(t *testing.T) {
		pool := NewMapPool()
		assert.NotNil(t, pool)
	})

	t.Run("Get and Put", func(t *testing.T) {
		pool := NewMapPool()

		// Get a map
		m := pool.Get()
		assert.NotNil(t, m)
		assert.IsType(t, map[string]interface{}{}, m)
		assert.Empty(t, m)

		// Add data to map
		m["key1"] = "value1"
		m["key2"] = 42
		assert.Len(t, m, 2)

		// Put map back
		pool.Put(m)

		// Map should be cleared
		assert.Empty(t, m)
	})

	t.Run("Multiple maps", func(t *testing.T) {
		pool := NewMapPool()

		m1 := pool.Get()
		m2 := pool.Get()

		m1["test1"] = "value1"
		m2["test2"] = "value2"

		assert.Equal(t, "value1", m1["test1"])
		assert.Equal(t, "value2", m2["test2"])

		pool.Put(m1)
		pool.Put(m2)

		// Get maps again - they should be empty
		m3 := pool.Get()
		m4 := pool.Get()

		assert.Empty(t, m3)
		assert.Empty(t, m4)
	})

	t.Run("Concurrent access", func(t *testing.T) {
		pool := NewMapPool()
		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				m := pool.Get()
				m["test"] = id
				assert.Equal(t, id, m["test"])
				pool.Put(m)
			}(i)
		}

		wg.Wait()
	})
}

func TestStringSlicePool(t *testing.T) {
	t.Run("NewStringSlicePool", func(t *testing.T) {
		pool := NewStringSlicePool()
		assert.NotNil(t, pool)
	})

	t.Run("Get and Put", func(t *testing.T) {
		pool := NewStringSlicePool()

		// Get a slice
		s := pool.Get()
		assert.NotNil(t, s)
		assert.IsType(t, []string{}, s)
		assert.Empty(t, s)
		assert.True(t, cap(s) >= 10) // Should have pre-allocated capacity

		// Add data to slice
		s = append(s, "item1", "item2", "item3")
		assert.Len(t, s, 3)

		// Put slice back
		pool.Put(s)

		// Slice should be reset to length 0 but keep capacity
		assert.Len(t, s, 0)
		assert.True(t, cap(s) >= 3) // Capacity should be preserved
	})

	t.Run("Multiple slices", func(t *testing.T) {
		pool := NewStringSlicePool()

		s1 := pool.Get()
		s2 := pool.Get()

		s1 = append(s1, "slice1")
		s2 = append(s2, "slice2")

		assert.Equal(t, []string{"slice1"}, s1)
		assert.Equal(t, []string{"slice2"}, s2)

		pool.Put(s1)
		pool.Put(s2)

		// Get slices again - they should be empty
		s3 := pool.Get()
		s4 := pool.Get()

		assert.Empty(t, s3)
		assert.Empty(t, s4)
	})

	t.Run("Concurrent access", func(t *testing.T) {
		pool := NewStringSlicePool()
		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				s := pool.Get()
				s = append(s, "test")
				assert.Equal(t, []string{"test"}, s)
				pool.Put(s)
			}(i)
		}

		wg.Wait()
	})
}

func TestGlobalPools(t *testing.T) {
	t.Run("Global pools exist", func(t *testing.T) {
		assert.NotNil(t, GlobalBufferPool)
		assert.NotNil(t, GlobalMapPool)
		assert.NotNil(t, GlobalStringSlicePool)
	})

	t.Run("GetBuffer and PutBuffer", func(t *testing.T) {
		buf := GetBuffer()
		assert.NotNil(t, buf)

		buf.WriteString("global test")
		assert.Equal(t, "global test", buf.String())

		PutBuffer(buf)
		assert.Equal(t, "", buf.String())
	})

	t.Run("GetMap and PutMap", func(t *testing.T) {
		m := GetMap()
		assert.NotNil(t, m)
		assert.Empty(t, m)

		m["global"] = "test"
		assert.Equal(t, "test", m["global"])

		PutMap(m)
		assert.Empty(t, m)
	})

	t.Run("GetStringSlice and PutStringSlice", func(t *testing.T) {
		s := GetStringSlice()
		assert.NotNil(t, s)
		assert.Empty(t, s)

		s = append(s, "global", "test")
		assert.Equal(t, []string{"global", "test"}, s)

		PutStringSlice(s)
		assert.Empty(t, s)
	})
}

func TestPoolReuse(t *testing.T) {
	t.Run("Buffer pool reuse", func(t *testing.T) {
		pool := NewBufferPool()

		// Get and put a buffer
		buf1 := pool.Get()
		buf1.WriteString("test")
		pool.Put(buf1)

		// Get another buffer - might be the same one
		buf2 := pool.Get()
		assert.Equal(t, "", buf2.String()) // Should be clean

		// They might be the same buffer object (pool reuse)
		// but we can't guarantee it due to sync.Pool behavior
	})

	t.Run("Map pool reuse", func(t *testing.T) {
		pool := NewMapPool()

		// Get and put a map
		m1 := pool.Get()
		m1["test"] = "value"
		pool.Put(m1)

		// Get another map - might be the same one
		m2 := pool.Get()
		assert.Empty(t, m2) // Should be clean
	})

	t.Run("String slice pool reuse", func(t *testing.T) {
		pool := NewStringSlicePool()

		// Get and put a slice
		s1 := pool.Get()
		s1 = append(s1, "test")
		pool.Put(s1)

		// Get another slice - might be the same one
		s2 := pool.Get()
		assert.Empty(t, s2) // Should be clean
	})
}

func TestPoolPerformance(t *testing.T) {
	t.Run("Buffer pool vs new allocation", func(t *testing.T) {
		pool := NewBufferPool()

		// Using pool
		for i := 0; i < 100; i++ {
			buf := pool.Get()
			buf.WriteString("test data")
			pool.Put(buf)
		}

		// Direct allocation (for comparison)
		for i := 0; i < 100; i++ {
			buf := &bytes.Buffer{}
			buf.WriteString("test data")
			// No explicit cleanup needed
		}

		// This test mainly ensures the pool works without errors
		// Performance comparison would require benchmarks
	})
}

func BenchmarkBufferPool(b *testing.B) {
	pool := NewBufferPool()

	b.Run("Pool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := pool.Get()
			buf.WriteString("benchmark test data")
			pool.Put(buf)
		}
	})

	b.Run("Direct", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := &bytes.Buffer{}
			buf.WriteString("benchmark test data")
		}
	})
}

func BenchmarkMapPool(b *testing.B) {
	pool := NewMapPool()

	b.Run("Pool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := pool.Get()
			m["key"] = "value"
			pool.Put(m)
		}
	})

	b.Run("Direct", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := make(map[string]interface{})
			m["key"] = "value"
		}
	})
}

func BenchmarkStringSlicePool(b *testing.B) {
	pool := NewStringSlicePool()

	b.Run("Pool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			s := pool.Get()
			s = append(s, "item")
			pool.Put(s)
		}
	})

	b.Run("Direct", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			s := make([]string, 0, 10)
			_ = append(s, "item")
		}
	})
}
