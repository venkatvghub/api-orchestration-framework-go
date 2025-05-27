package services

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/venkatvghub/api-orchestration-framework/pkg/config"
)

// CacheEntry represents a cached item with expiration
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// CacheService provides in-memory caching functionality
type CacheService struct {
	config *config.FrameworkConfig
	logger *zap.Logger
	cache  map[string]*CacheEntry
	mutex  sync.RWMutex
	ttl    time.Duration
}

// NewCacheService creates a new cache service
func NewCacheService(cfg *config.FrameworkConfig, logger *zap.Logger) *CacheService {
	service := &CacheService{
		config: cfg,
		logger: logger,
		cache:  make(map[string]*CacheEntry),
		ttl:    5 * time.Minute, // Default TTL
	}

	// Start cleanup goroutine
	go service.cleanup()

	return service
}

// Get retrieves an item from cache
func (s *CacheService) Get(key string) (interface{}, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entry, exists := s.cache[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		s.mutex.RUnlock()
		s.mutex.Lock()
		delete(s.cache, key)
		s.mutex.Unlock()
		s.mutex.RLock()
		return nil, false
	}

	s.logger.Debug("Cache hit", zap.String("key", key))
	return entry.Data, true
}

// Set stores an item in cache with default TTL
func (s *CacheService) Set(key string, value interface{}) {
	s.SetWithTTL(key, value, s.ttl)
}

// SetWithTTL stores an item in cache with custom TTL
func (s *CacheService) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.cache[key] = &CacheEntry{
		Data:      value,
		ExpiresAt: time.Now().Add(ttl),
	}

	s.logger.Debug("Cache set",
		zap.String("key", key),
		zap.Duration("ttl", ttl))
}

// Delete removes an item from cache
func (s *CacheService) Delete(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.cache, key)
	s.logger.Debug("Cache delete", zap.String("key", key))
}

// Clear removes all items from cache
func (s *CacheService) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.cache = make(map[string]*CacheEntry)
	s.logger.Info("Cache cleared")
}

// GetScreenCacheKey generates a cache key for screen data
func (s *CacheService) GetScreenCacheKey(screenID, deviceType string) string {
	return fmt.Sprintf("screen:%s:%s", screenID, deviceType)
}

// GetProgressCacheKey generates a cache key for user progress
func (s *CacheService) GetProgressCacheKey(userID string) string {
	return fmt.Sprintf("progress:%s", userID)
}

// GetAnalyticsCacheKey generates a cache key for analytics data
func (s *CacheService) GetAnalyticsCacheKey(userID string) string {
	return fmt.Sprintf("analytics:%s", userID)
}

// CacheScreen stores screen data in cache
func (s *CacheService) CacheScreen(screenID, deviceType string, screen interface{}) {
	key := s.GetScreenCacheKey(screenID, deviceType)
	s.Set(key, screen)
}

// GetCachedScreen retrieves screen data from cache
func (s *CacheService) GetCachedScreen(screenID, deviceType string) (interface{}, bool) {
	key := s.GetScreenCacheKey(screenID, deviceType)
	return s.Get(key)
}

// CacheProgress stores user progress in cache
func (s *CacheService) CacheProgress(userID string, progress interface{}) {
	key := s.GetProgressCacheKey(userID)
	s.SetWithTTL(key, progress, 2*time.Minute) // Shorter TTL for progress
}

// GetCachedProgress retrieves user progress from cache
func (s *CacheService) GetCachedProgress(userID string) (interface{}, bool) {
	key := s.GetProgressCacheKey(userID)
	return s.Get(key)
}

// CacheAnalytics stores analytics data in cache
func (s *CacheService) CacheAnalytics(userID string, analytics interface{}) {
	key := s.GetAnalyticsCacheKey(userID)
	s.SetWithTTL(key, analytics, 10*time.Minute) // Longer TTL for analytics
}

// GetCachedAnalytics retrieves analytics data from cache
func (s *CacheService) GetCachedAnalytics(userID string) (interface{}, bool) {
	key := s.GetAnalyticsCacheKey(userID)
	return s.Get(key)
}

// GetStats returns cache statistics
func (s *CacheService) GetStats() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	totalEntries := len(s.cache)
	expiredEntries := 0
	now := time.Now()

	for _, entry := range s.cache {
		if now.After(entry.ExpiresAt) {
			expiredEntries++
		}
	}

	return map[string]interface{}{
		"total_entries":   totalEntries,
		"expired_entries": expiredEntries,
		"active_entries":  totalEntries - expiredEntries,
		"default_ttl":     s.ttl.String(),
	}
}

// cleanup removes expired entries periodically
func (s *CacheService) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.Lock()
		now := time.Now()
		expiredKeys := []string{}

		for key, entry := range s.cache {
			if now.After(entry.ExpiresAt) {
				expiredKeys = append(expiredKeys, key)
			}
		}

		for _, key := range expiredKeys {
			delete(s.cache, key)
		}

		if len(expiredKeys) > 0 {
			s.logger.Debug("Cache cleanup completed",
				zap.Int("expired_entries", len(expiredKeys)))
		}

		s.mutex.Unlock()
	}
}

// MarshalJSON serializes cache entry for JSON output
func (e *CacheEntry) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"data":       e.Data,
		"expires_at": e.ExpiresAt,
	})
}
