package main

import (
	"testing"
)

func TestMruCache_BasicOps(t *testing.T) {
	cache, err := NewMRUCache[int, string](2)
	if err != nil {
		t.Fatalf("couldnt initialise the cache itself: %v", err)
	}
	cache.Put(1, "a")
	cache.Put(2, "b")
	if val, ok := cache.Get(2); !ok || val != "b" {
		t.Errorf("Expected b but got %v", val)
	}
	cache.Get(2)
	cache.Put(3, "c")
	if _, ok := cache.Get(2); ok {
		t.Errorf("expected 2 to be missing")
	}
	if val, ok := cache.Get(1); !ok || val != "a" {
		t.Errorf("expected 1 to remain but got %v", val)
	}
	if val, ok := cache.Get(3); !ok || val != "c" {
		t.Errorf("expected 3 to remain but got %v", val)
	}
}

func TestMRUCache_Eviction(t *testing.T) {
	tests := []struct {
		setup     func(c *MRUCache[int, string])
		capacity  int
		do        func(c *MRUCache[int, string])
		wantEvict map[int]string
	}{
		{
			setup: func(c *MRUCache[int, string]) {
				c.Put(1, "a")
				c.Put(2, "b")
			},
			do: func(c *MRUCache[int, string]) {
				c.Get(1)
				c.Put(3, "c")
			},
			capacity:  2,
			wantEvict: map[int]string{1: "a"},
		},
	}

	for _, tt := range tests {
		t.Run("test", func(t *testing.T) {
			c, _ := NewMRUCache[int, string](tt.capacity)
			tt.setup(c)
			tt.do(c)

			for k := range tt.wantEvict {
				if _, ok := c.Get(k); ok {
					t.Errorf("expected %d to be evicted but it's still there", k)
				}
			}

			for key, val := range map[int]string{2: "b", 3: "c"} {
				if got, ok := c.Get(key); !ok || got != val {
					t.Errorf("expected %d to be present with value %q but got %q", key, val, got)
				}
			}
		})
	}
}
