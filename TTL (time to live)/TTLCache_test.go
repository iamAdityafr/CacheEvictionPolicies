package main

import (
	"testing"
	"time"
)

func TestTTLCache_BasicOps(t *testing.T) {
	tests := []struct {
		name      string
		capacity  int
		setup     func(c *TTLCache[int, string])
		do        func(c *TTLCache[int, string])
		sleep     time.Duration
		wantEvict []struct {
			key   int
			value string
		}
	}{
		{
			name:     "basic ops",
			capacity: 2,
			setup: func(c *TTLCache[int, string]) {
				c.Set(1, "a", 1*time.Second)
				c.Set(2, "b", 1*time.Second)
			},
			do: func(c *TTLCache[int, string]) {
				c.Get(1)
				c.Get(2)
			},
			sleep: 2 * time.Second,
			wantEvict: []struct {
				key   int
				value string
			}{{1, "a"}, {2, "b"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewTTLCache[int, string]()
			if err != nil {
				t.Fatalf("couldnt initialise cache: %v", err)
			}
			tt.setup(c)
			tt.do(c)
			time.Sleep(tt.sleep)

			for _, v := range tt.wantEvict {
				got, ok := c.Get(v.key)
				if ok {
					t.Errorf("%d should have evicted but got %v", v.key, got)
				}
			}
		})
	}
}

func TestTTLCache_Eviction(t *testing.T) {
	tests := []struct {
		name  string
		setup func(c *TTLCache[int, string])
		sleep time.Duration
		check func(t *testing.T, c *TTLCache[int, string])
	}{
		{
			name: "expirign single",
			setup: func(c *TTLCache[int, string]) {
				c.Set(1, "a", 500*time.Millisecond)
			},
			sleep: 600 * time.Millisecond,
			check: func(t *testing.T, c *TTLCache[int, string]) {
				_, ok := c.Get(1)
				if ok {
					t.Errorf("item 1 should've expired")
				}
			},
		},
		{
			name: "multiple items from diff ttl",
			setup: func(c *TTLCache[int, string]) {
				c.Set(1, "a", 500*time.Millisecond)
				c.Set(2, "b", 1*time.Second)
			},
			sleep: 600 * time.Millisecond,
			check: func(t *testing.T, c *TTLCache[int, string]) {
				_, ok1 := c.Get(1)
				if ok1 {
					t.Errorf("item 1 should expired")
				}
				v2, ok2 := c.Get(2)
				if !ok2 || v2 != "b" {
					t.Errorf("item 2 should still exist got %v", v2)
				}
			},
		},
		{
			name: "overwrite",
			setup: func(c *TTLCache[int, string]) {
				c.Set(1, "a", 500*time.Millisecond)
				c.Set(1, "a2", 1*time.Second)
			},
			sleep: 600 * time.Millisecond,
			check: func(t *testing.T, c *TTLCache[int, string]) {
				v, ok := c.Get(1)
				if !ok || v != "a2" {
					t.Errorf("item 1 should have value a2 bur got %v", v)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := NewTTLCache[int, string]()
			tt.setup(c)
			time.Sleep(tt.sleep)
			tt.check(t, c)
			c.Stop()
		})
	}
}
