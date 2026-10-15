package engine

import (
	"math"
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

// ZeroOffset is the cron-equivalent offset: ticks land on clock boundaries
// with no jitter.
const ZeroOffset time.Duration = 0
