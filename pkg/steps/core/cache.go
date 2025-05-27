package core

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/venkatvghub/api-orchestration-framework/pkg/config"
	"github.com/venkatvghub/api-orchestration-framework/pkg/interfaces"
	"github.com/venkatvghub/api-orchestration-framework/pkg/metrics"
	"github.com/venkatvghub/api-orchestration-framework/pkg/utils"
	"go.uber.org/zap"
)

// CacheEntry represents a cached value with expiration
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
	CreatedAt time.Time
}

// IsExpired checks if the cache entry has expired
func (ce *CacheEntry) IsExpired() bool {
	return !ce.ExpiresAt.IsZero() && time.Now().After(ce.ExpiresAt)
}

// CacheStep provides thread-safe caching operations with TTL support
type CacheStep struct {
	name      string
	operation string // get, set, delete, clear
	key       string
	value     interface{}
	ttl       time.Duration
	saveAs    string

	// Thread-safe cache with TTL
	cache     sync.Map
	expiryMap sync.Map
	config    *config.CacheConfig
}

// NewCacheStep creates a new cache step
func NewCacheStep(name string) *CacheStep {
	return &CacheStep{
		name:   name,
		config: &config.DefaultConfig().Cache,
	}
}

// WithConfig sets the cache configuration
func (c *CacheStep) WithConfig(cfg *config.CacheConfig) *CacheStep {
	c.config = cfg
	return c
}

// Run executes the cache operation
func (c *CacheStep) Run(ctx interfaces.ExecutionContext) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		metrics.RecordStepExecution(c.Name(), duration, true)
		metrics.RecordCacheOperation(c.operation, false, duration) // Will be updated based on hit/miss
	}()

	// Interpolate key if it contains variables
	finalKey, err := utils.InterpolateString(c.key, ctx)
	if err != nil {
		return fmt.Errorf("key interpolation failed: %w", err)
	}

	switch c.operation {
	case "get":
		return c.handleGet(ctx, finalKey)
	case "set":
		return c.handleSet(ctx, finalKey)
	case "delete":
		return c.handleDelete(ctx, finalKey)
	case "clear":
		return c.handleClear(ctx)
	default:
		return fmt.Errorf("unsupported cache operation: %s", c.operation)
	}
}

func (cs *CacheStep) handleGet(ctx interfaces.ExecutionContext, cacheKey string) error {
	value, exists := cs.cache.Load(cacheKey)
	if !exists {
		ctx.Logger().Info("Cache miss",
			zap.String("step", cs.Name()),
			zap.String("cache_key", cacheKey))
		ctx.Set("cache_hit", false)
		return nil
	}

	entry, ok := value.(CacheEntry)
	if !ok {
		// Invalid cache entry, remove it
		cs.cache.Delete(cacheKey)
		ctx.Logger().Warn("Invalid cache entry removed",
			zap.String("step", cs.Name()),
			zap.String("cache_key", cacheKey))
		ctx.Set("cache_hit", false)
		return nil
	}

	if entry.IsExpired() {
		// Expired entry, remove it
		cs.cache.Delete(cacheKey)
		ctx.Logger().Info("Expired cache entry removed",
			zap.String("step", cs.Name()),
			zap.String("cache_key", cacheKey),
			zap.Time("expired_at", entry.ExpiresAt))
		ctx.Set("cache_hit", false)
		return nil
	}

	// Cache hit
	targetField := cs.saveAs
	if targetField == "" {
		targetField = "cached_value"
	}

	ctx.Set(targetField, entry.Value)
	ctx.Set("cache_hit", true)
	ctx.Set("cache_created_at", entry.CreatedAt)

	ctx.Logger().Info("Cache hit",
		zap.String("step", cs.Name()),
		zap.String("cache_key", cacheKey),
		zap.String("target_field", targetField),
		zap.Time("created_at", entry.CreatedAt))

	return nil
}

func (cs *CacheStep) handleSet(ctx interfaces.ExecutionContext, cacheKey string) error {
	// Get value to cache
	var valueToCache interface{}
	if cs.value != nil {
		valueToCache = cs.value
	} else {
		// Cache the entire context (excluding sensitive data)
		valueToCache = cs.sanitizeContextForCache(ctx)
	}

	// Create cache entry
	entry := CacheEntry{
		Value:     valueToCache,
		CreatedAt: time.Now(),
	}

	if cs.ttl > 0 {
		entry.ExpiresAt = time.Now().Add(cs.ttl)
	}

	// Store in cache
	cs.cache.Store(cacheKey, entry)

	ctx.Logger().Info("Value cached",
		zap.String("step", cs.Name()),
		zap.String("cache_key", cacheKey),
		zap.String("value_field", fmt.Sprintf("%v", cs.value)),
		zap.Duration("ttl", cs.ttl),
		zap.Time("expires_at", entry.ExpiresAt))

	return nil
}

func (cs *CacheStep) handleDelete(ctx interfaces.ExecutionContext, cacheKey string) error {
	_, existed := cs.cache.LoadAndDelete(cacheKey)

	ctx.Logger().Info("Cache entry deleted",
		zap.String("step", cs.Name()),
		zap.String("cache_key", cacheKey),
		zap.Bool("existed", existed))

	ctx.Set("cache_deleted", existed)
	return nil
}

func (cs *CacheStep) handleClear(ctx interfaces.ExecutionContext) error {
	count := 0
	cs.cache.Range(func(key, value interface{}) bool {
		cs.cache.Delete(key)
		count++
		return true
	})

	ctx.Logger().Info("Cache cleared",
		zap.String("step", cs.Name()),
		zap.Int("entries_cleared", count))

	ctx.Set("cache_cleared_count", count)
	return nil
}

func (cs *CacheStep) sanitizeContextForCache(ctx interfaces.ExecutionContext) map[string]interface{} {
	result := make(map[string]interface{})

	// Get all context keys and values
	for _, key := range ctx.Keys() {
		if value, ok := ctx.Get(key); ok {
			// Skip sensitive fields
			if cs.isSensitiveField(key) {
				continue
			}
			result[key] = value
		}
	}

	return result
}

func (cs *CacheStep) isSensitiveField(key string) bool {
	sensitiveFields := []string{
		"password", "token", "secret", "key", "auth",
		"authorization", "cookie", "session",
	}

	keyLower := strings.ToLower(key)
	for _, sensitive := range sensitiveFields {
		if strings.Contains(keyLower, sensitive) {
			return true
		}
	}

	return false
}

// Helper functions for creating common cache operations

// NewCacheGetStep creates a cache get step
func NewCacheGetStep(name, keyTemplate, targetField string) *CacheStep {
	return NewCacheStep(name).
		WithKey(keyTemplate).
		WithSaveAs(targetField)
}

// NewCacheSetStep creates a cache set step
func NewCacheSetStep(name, keyTemplate, valueField string, ttl time.Duration) *CacheStep {
	return NewCacheStep(name).
		WithKey(keyTemplate).
		WithValue(valueField).
		WithTTL(ttl)
}

// NewCacheDeleteStep creates a cache delete step
func NewCacheDeleteStep(name, keyTemplate string) *CacheStep {
	return NewCacheStep(name).
		WithKey(keyTemplate)
}

// NewCacheClearStep creates a cache clear step
func NewCacheClearStep(name string) *CacheStep {
	return NewCacheStep(name)
}

// CacheStats provides cache statistics
type CacheStats struct {
	TotalEntries   int
	ExpiredEntries int
	ValidEntries   int
}

// GetCacheStats returns statistics about the cache
func (cs *CacheStep) GetCacheStats() CacheStats {
	stats := CacheStats{}

	cs.cache.Range(func(key, value interface{}) bool {
		stats.TotalEntries++

		if entry, ok := value.(CacheEntry); ok {
			if entry.IsExpired() {
				stats.ExpiredEntries++
			} else {
				stats.ValidEntries++
			}
		}

		return true
	})

	return stats
}

// CleanupExpiredEntries removes expired entries from the cache
func (cs *CacheStep) CleanupExpiredEntries() int {
	removed := 0

	cs.cache.Range(func(key, value interface{}) bool {
		if entry, ok := value.(CacheEntry); ok && entry.IsExpired() {
			cs.cache.Delete(key)
			removed++
		}
		return true
	})

	return removed
}

// Name returns the step name
func (c *CacheStep) Name() string {
	return c.name
}

// Description returns the step description
func (c *CacheStep) Description() string {
	return fmt.Sprintf("Cache operation: %s", c.operation)
}

// WithOperation sets the cache operation
func (c *CacheStep) WithOperation(operation string) *CacheStep {
	c.operation = operation
	return c
}

// WithKey sets the cache key
func (c *CacheStep) WithKey(key string) *CacheStep {
	c.key = key
	return c
}

// WithValue sets the value to cache
func (c *CacheStep) WithValue(value interface{}) *CacheStep {
	c.value = value
	return c
}

// WithTTL sets the TTL for cache entries
func (c *CacheStep) WithTTL(ttl time.Duration) *CacheStep {
	c.ttl = ttl
	return c
}

// WithSaveAs sets the field name to save the result
func (c *CacheStep) WithSaveAs(saveAs string) *CacheStep {
	c.saveAs = saveAs
	return c
}
