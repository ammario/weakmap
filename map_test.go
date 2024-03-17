package weakmap

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTLRU(t *testing.T) {
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
