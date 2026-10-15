package engine

import (
	"math"
	"math/rand/v2"
	"time"

	"github.com/egdaemon/fron/internal/backoffx"
)

// seed supports any string or raw byte slice for deterministic hashing.
type seed interface {
	~string | ~[]byte
}

// Seed hashes a seed identifier into a deterministic offset.
func Seed[T seed](s T) time.Duration {
	return backoffx.DynamicHashDuration(time.Duration(math.MaxInt64), string(s))
}

// autoOffset resolves an auto seed id to an offset within one period.
func autoOffset(id string, hz time.Duration) time.Duration {
	if id == "" {
		return time.Duration(rand.Uint64() % uint64(hz))
	}

	return Seed(id) % hz
}
