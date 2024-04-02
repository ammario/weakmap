package weakmap

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	t.Run("OverrideValue", func(t *testing.T) {
		m := Map[string, int]{}
		m.Set("a", 10)
		m.Set("a", 20)
		v, ok := m.Get("a")
		if !ok {
			t.Fatalf("entry doesn't exist")
		}
		if v != 20 {
			t.Fatalf("v is %v", v)
		}
	})

	t.Run("DeleteEntry", func(t *testing.T) {
		m := Map[string, int]{}
		m.Set("a", 10)
		m.Delete("a")
		v, ok := m.Get("a")
		if ok {
			t.Fatalf("value %v:%v still exists", "a", v)
		}
	})

	t.Run("Do", func(t *testing.T) {
		c := Map[string, int]{}

		n := 10
		fn := func() (int, error) {
			n += 1
			return n, nil
		}

		v, err := c.Do("a", fn)
		require.NoError(t, err)

		require.Equal(t, 11, v)

		v, err = c.Do("a", fn)
		require.NoError(t, err)

		// No recompute, cache hit.
		require.Equal(t, 11, v)
	})
}

func TestMap_Cost(t *testing.T) {
	t.Run("OldValuesEvicted", func(t *testing.T) {
		m := Map[string, int]{MaxCost: 10}
		for i := 0; i < 100; i++ {
			m.Set(strconv.Itoa(i), i)
			// 4 is our busy value that should not be evicted.
			m.Get("4")
		}
		for i := 0; i < 100; i++ {
			v, ok := m.Get(strconv.Itoa(i))
			if i < 91 && i != 4 {
				if ok {
					t.Fatalf("value %v:%v exists", i, v)
				}
				continue
			}
			if !ok {
				t.Fatalf("value %v:%v should be in cache", i, v)
			}
			if m.cost != 10 {
				t.Fatalf("cost is %v", m.cost)
			}
			if len(m.index) != 10 {
				t.Fatalf("len(c.index) is %v", len(m.index))
			}
		}
	})
	t.Run("LimitBytes", func(t *testing.T) {
		m := Map[string, []byte]{
			Coster:  func(v []byte) int { return len(v) },
			MaxCost: 100,
		}

		const allocSize = 5
		for i := 0; i < 100; i++ {
			m.Set("big"+strconv.Itoa(i), make([]byte, allocSize))
		}

		if m.Cost() > m.MaxCost {
			t.Fatalf("cost is %v", m.Cost())
		}

		if wantLen := m.MaxCost / allocSize; m.Len() != wantLen {
			t.Fatalf("len is %v, want %v", m.Len(), wantLen)
		}
	})
}

func BenchmarkGet(b *testing.B) {
	m := Map[string, int]{}
	m.Set("test-key", 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get("test-key")
	}
}

func BenchmarkSet(b *testing.B) {
	m := Map[string, []byte]{}
	const allocSize = 1024
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		byt := make([]byte, allocSize)
		b.StartTimer()
		m.Set("test-key-"+strconv.Itoa(i), byt)
	}
}
