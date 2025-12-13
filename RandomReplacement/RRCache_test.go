package main

import (
	"strconv"
	"testing"
)

func TestRandomCache_Eviction(t *testing.T) {
	tests := []struct {
		capacity    int
		insertCount int
	}{
		{3, 4},
		{5, 10},
	}

	for i, tt := range tests {
		t.Run("test: "+strconv.Itoa(i), func(t *testing.T) {
			c, _ := NewRandomCache[int, string](tt.capacity)

			for i := 0; i < tt.insertCount; i++ {
				c.Put(i, string(rune('a'+i)))
			}

			if got := c.Len(); got > tt.capacity {
				t.Error("cache exceeded capacity and got: %d", got)
			}

			for i := 0; i < tt.insertCount-tt.capacity; i++ {
				if _, ok := c.Get(i); ok {
					t.Errorf("%d should have evicted but is still there", i)
				}
			}
		})
	}
}
