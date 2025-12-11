package main

import "testing"

func TestLFUCache_BasicOperations(t *testing.T) {
	cache, err := NewLFUCache[int, string](2)
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}
	cache.Put(1, "A")
	cache.Put(2, "B")
	if val, ok := cache.Get(1); !ok || val != "A" {
		t.Errorf("expected 1 to get A but got %v", val)
	}
	cache.Put(1, "C")
	if val, _ := cache.Get(1); val != "C" {
		t.Errorf("expected 1 to get updated to C but got %v", val)
	}

	if _, ok := cache.Get(3); ok {
		t.Errorf("expected 3 to be missing")
	}
}
func TestLFUCache_ZeroCapacity(t *testing.T) {
	_, err := NewLFUCache[int, string](0)
	if err == nil {
		t.Fatal("expected error for zero capacity")
	}
}
func TestLFUCache_CapacityOne(t *testing.T) {
	cache, _ := NewLFUCache[int, string](1)
	cache.Put(1, "A")
	if value, ok := cache.Get(1); !ok || value != "A" {
		t.Errorf("Wanted A but got %v", value)
	}
	cache.Put(2, "B")
	if _, ok := cache.Get(1); ok {
		t.Errorf("1 now evicted")
	}
	if value, ok := cache.Get(2); !ok || value != "B" {
		t.Errorf("Expecting B but got %v", value)
	}
}
func TestLFUCache_Eviction(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(c *LFUCache[int, string])
		wantEvict map[int]string
		capacity  int
	}{
		{
			name:     "evict least frequently used",
			capacity: 2,
			setup: func(c *LFUCache[int, string]) {
				c.Put(1, "A")
				c.Put(2, "B")
				c.Get(1)
				c.Put(3, "C")
			},
			wantEvict: map[int]string{2: "B"},
		}, {

			name:     "break lru test",
			capacity: 2,
			setup: func(c *LFUCache[int, string]) {
				c.Put(1, "A")
				c.Put(2, "B")
				c.Put(3, "C")
			},
			wantEvict: map[int]string{1: "A"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, _ := NewLFUCache[int, string](tt.capacity)
			tt.setup(cache)
			for k := range tt.wantEvict {
				if v, ok := cache.Get(k); ok {
					t.Errorf("expected %v to be evicted but got: %v", k, v)
				}
			}
		})
	}
}
