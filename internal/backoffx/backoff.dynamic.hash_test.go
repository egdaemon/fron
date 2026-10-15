package backoffx_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/egdaemon/fron/internal/backoffx"
	"github.com/stretchr/testify/require"
)

func TestDynamicHashWindow(t *testing.T) {
	t.Run("should be deterministic", func(t *testing.T) {
		for range 10 {
			require.Equal(t, backoffx.DynamicHashWindow("foo", 1000), backoffx.DynamicHashWindow("foo", 1000))
		}
	})

	t.Run("should remain within the window", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			require.Less(t, backoffx.DynamicHashWindow(fmt.Sprintf("seed-%d", i), 17), uint64(17))
		}
	})

	t.Run("should be zero for a window of one", func(t *testing.T) {
		require.Equal(t, uint64(0), backoffx.DynamicHashWindow("foo", 1))
	})

	t.Run("should spread different seeds across the window", func(t *testing.T) {
		seen := make(map[uint64]struct{})
		for i := 0; i < 1000; i++ {
			seen[backoffx.DynamicHashWindow(fmt.Sprintf("seed-%d", i), 10)] = struct{}{}
		}
		require.Len(t, seen, 10)
	})

	t.Run("should remain compatible with the md5 little endian derivation", func(t *testing.T) {
		// md5("foo") = acbd18db4cc2f85cedef654fccc4a4d8, first 8 bytes little endian.
		require.Equal(t, uint64(0x5cf8c24cdb18bdac), backoffx.DynamicHashWindow("foo", math.MaxUint64))
	})
}

func TestDynamicHashDuration(t *testing.T) {
	t.Run("should be zero for an empty window", func(t *testing.T) {
		require.Equal(t, time.Duration(0), backoffx.DynamicHashDuration(0, "foo"))
		require.Equal(t, time.Duration(0), backoffx.DynamicHashDuration(-time.Minute, "foo"))
	})

	t.Run("should be deterministic and within the window", func(t *testing.T) {
		d := backoffx.DynamicHashDuration(time.Minute, "foo")
		require.Equal(t, d, backoffx.DynamicHashDuration(time.Minute, "foo"))
		require.GreaterOrEqual(t, d, time.Duration(0))
		require.Less(t, d, time.Minute)
	})
}
