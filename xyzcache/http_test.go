package xyzcache

import (
	"fmt"
	"log"
	"net/http"
	"testing"
)

func TestHTTPPool_ServeHTTP(t *testing.T) {
	NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9999"
	peers := NewHTTPPool(addr)
	log.Println("xyzcache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
