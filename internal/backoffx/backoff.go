package backoffx

import (
	"crypto/md5"
	"encoding/binary"
	"time"
)

// generate a *consistent* duration based on the input i within the
// provided window.
func DynamicHashDuration(window time.Duration, i string) time.Duration {
	if window <= 0 {
		return 0
	}

	return time.Duration(DynamicHashWindow(i, uint64(window)))
}

// uint64 to prevent negative values
func DynamicHashWindow(i string, n uint64) uint64 {
	digest := md5.Sum([]byte(i))
	return binary.LittleEndian.Uint64(digest[:]) % n
}
