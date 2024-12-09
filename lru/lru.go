package lru

import (
	"container/list"
)

// Cache is a LRU cache
type Cache struct {
	maxBytes int64
	nbytes   int64
	ll       *list.List
	cache    map[string]*list.Element

	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func NewLRUCache(maxBytes int64, OnEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		OnEvicted: OnEvicted,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
	}
}

// Get 查
func (c *Cache) Get(key string) (Value, bool) {
	//cache中有缓存，断言成entry后直接返回
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	//没有找到
	return nil, false
}

func (c *Cache) RemoveOldest() {
	if c.ll != nil {
		ele := c.ll.Back()
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(kv.value.Len()) + int64(len(kv.key))
		c.ll.Remove(ele)

		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add 存在就修改，不存在就增加
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key: key, value: value})
		c.cache[key] = ele
		c.nbytes += int64(value.Len()) + int64(len(key))
	}

	for c.maxBytes != 0 && c.nbytes > c.maxBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
