package core

import (
	"testing"
	"time"

	"github.com/venkatvghub/api-orchestration-framework/pkg/config"
	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
	"go.uber.org/zap"
)

func TestCacheEntry_IsExpired(t *testing.T) {
	// Test non-expired entry
	entry := CacheEntry{
		Value:     "test_value",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	if entry.IsExpired() {
		t.Error("Entry should not be expired")
	}

	// Test expired entry
	expiredEntry := CacheEntry{
		Value:     "test_value",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}

	if !expiredEntry.IsExpired() {
		t.Error("Entry should be expired")
	}

	// Test entry with zero expiration (never expires)
	neverExpiresEntry := CacheEntry{
		Value:     "test_value",
		ExpiresAt: time.Time{},
		CreatedAt: time.Now(),
	}

	if neverExpiresEntry.IsExpired() {
		t.Error("Entry with zero expiration should never expire")
	}
}

func TestNewCacheStep(t *testing.T) {
	step := NewCacheStep("test_cache")

	if step.Name() != "test_cache" {
		t.Errorf("Name() = %v, want 'test_cache'", step.Name())
	}

	if step.Description() != "Cache operation: " {
		t.Errorf("Description() = %v, want 'Cache operation: '", step.Description())
	}

	if step.config == nil {
		t.Error("Config should be initialized with default")
	}
}

func TestCacheStep_WithConfig(t *testing.T) {
	step := NewCacheStep("test_cache")
	cfg := &config.CacheConfig{}

	result := step.WithConfig(cfg)

	if result != step {
		t.Error("WithConfig should return the same instance")
	}

	if step.config != cfg {
		t.Error("Config should be set")
	}
}

func TestCacheStep_WithOperation(t *testing.T) {
	step := NewCacheStep("test_cache")
	result := step.WithOperation("get")

	if result != step {
		t.Error("WithOperation should return the same instance")
	}

	if step.operation != "get" {
		t.Errorf("operation = %v, want 'get'", step.operation)
	}
}

func TestCacheStep_WithKey(t *testing.T) {
	step := NewCacheStep("test_cache")
	result := step.WithKey("test_key")

	if result != step {
		t.Error("WithKey should return the same instance")
	}

	if step.key != "test_key" {
		t.Errorf("key = %v, want 'test_key'", step.key)
	}
}

func TestCacheStep_WithValue(t *testing.T) {
	step := NewCacheStep("test_cache")
	result := step.WithValue("test_value")

	if result != step {
		t.Error("WithValue should return the same instance")
	}

	if step.value != "test_value" {
		t.Errorf("value = %v, want 'test_value'", step.value)
	}
}

func TestCacheStep_WithTTL(t *testing.T) {
	step := NewCacheStep("test_cache")
	ttl := 5 * time.Minute
	result := step.WithTTL(ttl)

	if result != step {
		t.Error("WithTTL should return the same instance")
	}

	if step.ttl != ttl {
		t.Errorf("ttl = %v, want %v", step.ttl, ttl)
	}
}

func TestCacheStep_WithSaveAs(t *testing.T) {
	step := NewCacheStep("test_cache")
	result := step.WithSaveAs("result_field")

	if result != step {
		t.Error("WithSaveAs should return the same instance")
	}

	if step.saveAs != "result_field" {
		t.Errorf("saveAs = %v, want 'result_field'", step.saveAs)
	}
}

func TestCacheStep_Run_UnsupportedOperation(t *testing.T) {
	step := NewCacheStep("test_cache").
		WithOperation("invalid_operation")

	ctx := flow.NewContext()
	err := step.Run(ctx)

	if err == nil {
		t.Error("Run() should fail with unsupported operation")
	}

	if err.Error() != "unsupported cache operation: invalid_operation" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestCacheStep_HandleSet(t *testing.T) {
	step := NewCacheStep("test_cache").
		WithOperation("set").
		WithKey("test_key").
		WithValue("test_value").
		WithTTL(5 * time.Minute)

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	// Verify the value was cached
	value, exists := step.cache.Load("test_key")
	if !exists {
		t.Error("Value should be cached")
	}

	entry, ok := value.(CacheEntry)
	if !ok {
		t.Error("Cached value should be a CacheEntry")
	}

	if entry.Value != "test_value" {
		t.Errorf("Cached value = %v, want 'test_value'", entry.Value)
	}

	if entry.IsExpired() {
		t.Error("Entry should not be expired immediately after caching")
	}
}

func TestCacheStep_HandleGet_Hit(t *testing.T) {
	step := NewCacheStep("test_cache").
		WithOperation("get").
		WithKey("test_key").
		WithSaveAs("cached_result")

	// Pre-populate cache
	entry := CacheEntry{
		Value:     "cached_value",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}
	step.cache.Store("test_key", entry)

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	// Check cache hit
	cacheHit, ok := ctx.Get("cache_hit")
	if !ok || cacheHit != true {
		t.Error("cache_hit should be true")
	}

	// Check retrieved value
	result, ok := ctx.Get("cached_result")
	if !ok || result != "cached_value" {
		t.Error("Cached value should be retrieved")
	}

	// Check created_at timestamp
	createdAt, ok := ctx.Get("cache_created_at")
	if !ok {
		t.Error("cache_created_at should be set")
	}

	if _, ok := createdAt.(time.Time); !ok {
		t.Error("cache_created_at should be a time.Time")
	}
}

func TestCacheStep_HandleGet_Miss(t *testing.T) {
	step := NewCacheStep("test_cache").
		WithOperation("get").
		WithKey("nonexistent_key").
		WithSaveAs("cached_result")

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	// Check cache miss
	cacheHit, ok := ctx.Get("cache_hit")
	if !ok || cacheHit != false {
		t.Error("cache_hit should be false")
	}

	// Check that no value was retrieved
	_, ok = ctx.Get("cached_result")
	if ok {
		t.Error("No value should be retrieved on cache miss")
	}
}

func TestCacheStep_HandleGet_Expired(t *testing.T) {
	step := NewCacheStep("test_cache").
		WithOperation("get").
		WithKey("expired_key").
		WithSaveAs("cached_result")

	// Pre-populate cache with expired entry
	expiredEntry := CacheEntry{
		Value:     "expired_value",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}
	step.cache.Store("expired_key", expiredEntry)

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	// Check cache miss (expired entry should be treated as miss)
	cacheHit, ok := ctx.Get("cache_hit")
	if !ok || cacheHit != false {
		t.Error("cache_hit should be false for expired entry")
	}

	// Verify expired entry was removed
	_, exists := step.cache.Load("expired_key")
	if exists {
		t.Error("Expired entry should be removed from cache")
	}
}

func TestCacheStep_HandleGet_InvalidEntry(t *testing.T) {
	step := NewCacheStep("test_cache").
		WithOperation("get").
		WithKey("invalid_key").
		WithSaveAs("cached_result")

	// Pre-populate cache with invalid entry (not a CacheEntry)
	step.cache.Store("invalid_key", "invalid_entry")

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	// Check cache miss (invalid entry should be treated as miss)
	cacheHit, ok := ctx.Get("cache_hit")
	if !ok || cacheHit != false {
		t.Error("cache_hit should be false for invalid entry")
	}

	// Verify invalid entry was removed
	_, exists := step.cache.Load("invalid_key")
	if exists {
		t.Error("Invalid entry should be removed from cache")
	}
}

func TestCacheStep_HandleDelete(t *testing.T) {
	step := NewCacheStep("test_cache").
		WithOperation("delete").
		WithKey("delete_key")

	// Pre-populate cache
	entry := CacheEntry{
		Value:     "to_delete",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}
	step.cache.Store("delete_key", entry)

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	// Check that entry was deleted
	_, exists := step.cache.Load("delete_key")
	if exists {
		t.Error("Entry should be deleted from cache")
	}

	// Check deletion flag
	deleted, ok := ctx.Get("cache_deleted")
	if !ok || deleted != true {
		t.Error("cache_deleted should be true")
	}
}

func TestCacheStep_HandleDelete_NonExistent(t *testing.T) {
	step := NewCacheStep("test_cache").
		WithOperation("delete").
		WithKey("nonexistent_key")

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	// Check deletion flag
	deleted, ok := ctx.Get("cache_deleted")
	if !ok || deleted != false {
		t.Error("cache_deleted should be false for non-existent key")
	}
}

func TestCacheStep_HandleClear(t *testing.T) {
	step := NewCacheStep("test_cache").
		WithOperation("clear")

	// Pre-populate cache with multiple entries
	entries := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for key, value := range entries {
		entry := CacheEntry{
			Value:     value,
			ExpiresAt: time.Now().Add(1 * time.Hour),
			CreatedAt: time.Now(),
		}
		step.cache.Store(key, entry)
	}

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	// Check that all entries were cleared
	for key := range entries {
		_, exists := step.cache.Load(key)
		if exists {
			t.Errorf("Entry %s should be cleared from cache", key)
		}
	}

	// Check cleared count
	clearedCount, ok := ctx.Get("cache_cleared_count")
	if !ok || clearedCount != 3 {
		t.Errorf("cache_cleared_count should be 3, got %v", clearedCount)
	}
}

func TestCacheStep_HandleSet_WithContextValue(t *testing.T) {
	step := NewCacheStep("test_cache").
		WithOperation("set").
		WithKey("context_key").
		WithTTL(5 * time.Minute)
	// No explicit value set, should cache entire context

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	ctx.Set("test_field", "test_value")
	ctx.Set("another_field", 42)

	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	// Verify the context was cached
	value, exists := step.cache.Load("context_key")
	if !exists {
		t.Error("Context should be cached")
	}

	entry, ok := value.(CacheEntry)
	if !ok {
		t.Error("Cached value should be a CacheEntry")
	}

	contextMap, ok := entry.Value.(map[string]interface{})
	if !ok {
		t.Error("Cached context should be a map")
	}

	if contextMap["test_field"] != "test_value" {
		t.Error("Context field should be cached")
	}

	if contextMap["another_field"] != 42 {
		t.Error("Context field should be cached")
	}
}

func TestCacheStep_SanitizeContextForCache(t *testing.T) {
	step := NewCacheStep("test_cache")

	ctx := flow.NewContext()
	ctx.Set("safe_field", "safe_value")
	ctx.Set("password", "secret123")
	ctx.Set("auth_token", "token123")
	ctx.Set("api_key", "key123")
	ctx.Set("normal_field", "normal_value")

	sanitized := step.sanitizeContextForCache(ctx)

	// Safe fields should be included
	if sanitized["safe_field"] != "safe_value" {
		t.Error("Safe field should be included")
	}

	if sanitized["normal_field"] != "normal_value" {
		t.Error("Normal field should be included")
	}

	// Sensitive fields should be excluded
	if _, exists := sanitized["password"]; exists {
		t.Error("Password field should be excluded")
	}

	if _, exists := sanitized["auth_token"]; exists {
		t.Error("Auth token field should be excluded")
	}

	if _, exists := sanitized["api_key"]; exists {
		t.Error("API key field should be excluded")
	}
}

func TestCacheStep_IsSensitiveField(t *testing.T) {
	step := NewCacheStep("test_cache")

	sensitiveFields := []string{
		"password", "PASSWORD", "user_password",
		"token", "TOKEN", "auth_token",
		"secret", "SECRET", "api_secret",
		"key", "KEY", "api_key",
		"auth", "AUTH", "authorization",
		"cookie", "COOKIE", "session_cookie",
		"session", "SESSION", "user_session",
	}

	for _, field := range sensitiveFields {
		if !step.isSensitiveField(field) {
			t.Errorf("Field %s should be considered sensitive", field)
		}
	}

	safeFields := []string{
		"username", "email", "name", "id",
		"data", "result", "response", "request",
	}

	for _, field := range safeFields {
		if step.isSensitiveField(field) {
			t.Errorf("Field %s should not be considered sensitive", field)
		}
	}
}

func TestCacheStep_GetCacheStats(t *testing.T) {
	step := NewCacheStep("test_cache")

	// Add some entries
	validEntry := CacheEntry{
		Value:     "valid_value",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}
	step.cache.Store("valid_key", validEntry)

	expiredEntry := CacheEntry{
		Value:     "expired_value",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}
	step.cache.Store("expired_key", expiredEntry)

	neverExpiresEntry := CacheEntry{
		Value:     "never_expires",
		ExpiresAt: time.Time{},
		CreatedAt: time.Now(),
	}
	step.cache.Store("never_expires_key", neverExpiresEntry)

	stats := step.GetCacheStats()

	if stats.TotalEntries != 3 {
		t.Errorf("TotalEntries = %d, want 3", stats.TotalEntries)
	}

	if stats.ExpiredEntries != 1 {
		t.Errorf("ExpiredEntries = %d, want 1", stats.ExpiredEntries)
	}

	if stats.ValidEntries != 2 {
		t.Errorf("ValidEntries = %d, want 2", stats.ValidEntries)
	}
}

func TestCacheStep_CleanupExpiredEntries(t *testing.T) {
	step := NewCacheStep("test_cache")

	// Add some entries
	validEntry := CacheEntry{
		Value:     "valid_value",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}
	step.cache.Store("valid_key", validEntry)

	expiredEntry1 := CacheEntry{
		Value:     "expired_value1",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}
	step.cache.Store("expired_key1", expiredEntry1)

	expiredEntry2 := CacheEntry{
		Value:     "expired_value2",
		ExpiresAt: time.Now().Add(-30 * time.Minute),
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}
	step.cache.Store("expired_key2", expiredEntry2)

	removed := step.CleanupExpiredEntries()

	if removed != 2 {
		t.Errorf("CleanupExpiredEntries() = %d, want 2", removed)
	}

	// Valid entry should still exist
	_, exists := step.cache.Load("valid_key")
	if !exists {
		t.Error("Valid entry should still exist")
	}

	// Expired entries should be removed
	_, exists = step.cache.Load("expired_key1")
	if exists {
		t.Error("Expired entry 1 should be removed")
	}

	_, exists = step.cache.Load("expired_key2")
	if exists {
		t.Error("Expired entry 2 should be removed")
	}
}

// Test helper functions

func TestNewCacheGetStep(t *testing.T) {
	step := NewCacheGetStep("get_test", "key_${id}", "result_field")

	if step.Name() != "get_test" {
		t.Errorf("Name() = %v, want 'get_test'", step.Name())
	}

	// Note: Helper functions don't set operation, that's done separately
	if step.key != "key_${id}" {
		t.Errorf("key = %v, want 'key_${id}'", step.key)
	}

	if step.saveAs != "result_field" {
		t.Errorf("saveAs = %v, want 'result_field'", step.saveAs)
	}
}

func TestNewCacheSetStep(t *testing.T) {
	ttl := 10 * time.Minute
	step := NewCacheSetStep("set_test", "key_${id}", "value_field", ttl)

	if step.Name() != "set_test" {
		t.Errorf("Name() = %v, want 'set_test'", step.Name())
	}

	// Note: Helper functions don't set operation, that's done separately
	if step.key != "key_${id}" {
		t.Errorf("key = %v, want 'key_${id}'", step.key)
	}

	if step.value != "value_field" {
		t.Errorf("value = %v, want 'value_field'", step.value)
	}

	if step.ttl != ttl {
		t.Errorf("ttl = %v, want %v", step.ttl, ttl)
	}
}

func TestNewCacheDeleteStep(t *testing.T) {
	step := NewCacheDeleteStep("delete_test", "key_${id}")

	if step.Name() != "delete_test" {
		t.Errorf("Name() = %v, want 'delete_test'", step.Name())
	}

	// Note: Helper functions don't set operation, that's done separately
	if step.key != "key_${id}" {
		t.Errorf("key = %v, want 'key_${id}'", step.key)
	}
}

func TestNewCacheClearStep(t *testing.T) {
	step := NewCacheClearStep("clear_test")

	if step.Name() != "clear_test" {
		t.Errorf("Name() = %v, want 'clear_test'", step.Name())
	}

	// Note: Helper functions don't set operation, that's done separately
}
