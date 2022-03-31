//封装lru，以实现并发读写的存储
package main

import (
	"DistributedCache/pkg/lru"
	"sync"
)

type cache struct {
	mu               sync.Mutex
	lru              *lru.Cache
	setCacheMaxBytes int64 //设置的最大空间
}

func (c *cache) add(key string, value lru.Value) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.NewCache(c.setCacheMaxBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (ByteView, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		//延迟初始化
		return ByteView{}, false
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), true
	}
	return ByteView{}, false
}
