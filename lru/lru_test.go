package lru

import (
	"github.com/go-playground/assert/v2"
	"testing"
)

type stringX string

func (s stringX) Len() int {
	return len(s)
}

func TestGet(t *testing.T) {
	lru := NewLRUCache(int64(5000), nil)
	lru.Add("key1", stringX("1234"))
	if v, ok := lru.Get("key1"); !ok || string(v.(stringX)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "v1", "v2", "v3"

	cap := len(k1 + v1 + k2 + v2)

	lru := NewLRUCache(int64(cap), nil)
	lru.Add(k1, stringX(v1))
	lru.Add(k2, stringX(v2))
	lru.Add(k3, stringX(v3))

	if _, ok := lru.Get(k1); ok || lru.Len() != 2 {
		t.Fatalf("remove failed")
	}
}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	lru := NewLRUCache(int64(10), callback)
	lru.Add("key1", stringX("123456"))
	lru.Add("k2", stringX("k2"))
	lru.Add("k3", stringX("k3"))
	lru.Add("k4", stringX("k4"))

	expect := []string{"key1", "k2"}

	assert.Equal(t, expect, keys)
}
