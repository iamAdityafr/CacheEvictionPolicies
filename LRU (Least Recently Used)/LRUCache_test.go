package main

import "testing"

func TestLRUCache_BasicOps(t *testing.T) {
	tests := []struct {
		key   []int
		expt  []string
		found []bool
		setup func(c *LRUCache[int, string])
	}{
		{
			key:   []int{1, 2, 3},
			expt:  []string{"c", "b", ""},
			found: []bool{true, true, false},
			setup: func(c *LRUCache[int, string]) {
				c.Put(1, "a")
				c.Put(2, "b")
				c.Put(1, "c")
			},
		},
	}

	for _, tt := range tests {
		t.Run("basic ops", func(t *testing.T) {
			c, _ := NewLRUCache[int, string](2)
			tt.setup(c)
			for i, k := range tt.key {
				v, ok := c.Get(k)
				if ok != tt.found[i] {
					t.Errorf("expected found: %v but got %v", tt.found[i], k)
				}
				if ok && v != tt.expt[i] {
					t.Errorf("expected value: %q but got %q", tt.expt[i], v)
				}
			}
		})
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(c *LRUCache[int, string])
		do        func(c *LRUCache[int, string])
		capacity  int
		wantEvict map[int]string
	}{
		{
			name: "test 1",
			setup: func(c *LRUCache[int, string]) {
				c.Put(1, "a")
				c.Put(2, "b")
				c.Put(3, "c")
			},
			do: func(c *LRUCache[int, string]) {
				c.Get(2)
				c.Get(3)
			},
			capacity:  2,
			wantEvict: map[int]string{1: "a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := NewLRUCache[int, string](tt.capacity)
			tt.setup(c)
			tt.do(c)
			for k := range tt.wantEvict {
				if _, ok := c.Get(k); ok {
					t.Errorf("expected %d to be evicted but still present", k)
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
