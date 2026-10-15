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

// ZeroOffsetSeed is the reserved seed selecting the cron-equivalent zero offset.
const ZeroOffsetSeed = "fron.zero.offset"

// autoOffset resolves an auto seed id to an offset within one period.
func autoOffset(id string, hz time.Duration) time.Duration {
	switch id {
	case ZeroOffsetSeed:
		return 0
	case "":
		return time.Duration(rand.Uint64() % uint64(hz))
	default:
		return Seed(id) % hz
	}
}

// ZeroOffset is the cron-equivalent offset: ticks land on clock boundaries
// with no jitter.
const ZeroOffset time.Duration = 0
