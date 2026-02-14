package cache

import (
	"sync"
	"time"
)

// CacheItem 缓存项结构
type CacheItem struct {
	Value      interface{}
	ExpireTime time.Time
}

// IsExpired 检查缓存项是否过期
func (item *CacheItem) IsExpired() bool {
	return time.Now().After(item.ExpireTime)
}

// LocalCache 本地内存缓存
type LocalCache struct {
	mu      sync.RWMutex
	items   map[string]*CacheItem
	maxSize int
}

// NewLocalCache 创建新的本地缓存实例
// maxSize: 最大缓存条目数，0表示无限制
func NewLocalCache(maxSize int) *LocalCache {
	return &LocalCache{
		items:   make(map[string]*CacheItem),
		maxSize: maxSize,
	}
}

// Set 设置缓存项
// key: 缓存键
// value: 缓存值
// ttl: 过期时间
func (c *LocalCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果达到最大容量，清理过期项
	if c.maxSize > 0 && len(c.items) >= c.maxSize {
		c.cleanExpired()
	}

	c.items[key] = &CacheItem{
		Value:      value,
		ExpireTime: time.Now().Add(ttl),
	}
}

// Get 获取缓存项
// 返回值和是否存在标志
func (c *LocalCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// 检查是否过期
	if item.IsExpired() {
		c.Delete(key)
		return nil, false
	}

	return item.Value, true
}

// Delete 删除缓存项
func (c *LocalCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// DeleteByPrefix 根据前缀删除缓存项
func (c *LocalCache) DeleteByPrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.items {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.items, key)
		}
	}
}

// cleanExpired 清理过期缓存项
func (c *LocalCache) cleanExpired() {
	now := time.Now()
	for key, item := range c.items {
		if now.After(item.ExpireTime) {
			delete(c.items, key)
		}
	}
}

// Size 获取当前缓存数量
func (c *LocalCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Clear 清空所有缓存
func (c *LocalCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*CacheItem)
}

// 默认缓存实例（包级别共享）
var defaultCache = NewLocalCache(DefaultCacheSize)

// Default 获取默认缓存实例
func Default() *LocalCache {
	return defaultCache
}
