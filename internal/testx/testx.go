package testx

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Must panics when err is non-nil, otherwise returns v.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}

func Fixture(path ...string) string {
	return filepath.Join(".fixtures", filepath.Join(path...))
}

// Read a file at the given path.
func Read(path ...string) io.Reader {
	return bytes.NewReader(Must(os.ReadFile(filepath.Join(path...))))
}

// ReadString from the given file.
func ReadString(path ...string) string {
	return string(Must(os.ReadFile(filepath.Join(path...))))
}

// MustT is a small language extension for the common value, error return pattern.
func MustT[T any](v T, err error) func(t testing.TB) T {
	return func(t testing.TB) T {
		require.NoError(t, err)
		return v
	}
}

// Tempenvvar temporarily set the environment variable.
func Tempenvvar(k, v string, do func()) {
	o := os.Getenv(k)
	defer os.Setenv(k, o)
	if err := os.Setenv(k, v); err != nil {
		panic(err)
	}
	do()
}

func IOString(in io.Reader) string {
	return string(Must(io.ReadAll(in)))
}

func IOBytes(in io.Reader) []byte {
	return Must(io.ReadAll(in))
}

func Context(t *testing.T) (context.Context, context.CancelFunc) {
	if deadline, ok := t.Deadline(); ok {
		return context.WithDeadline(context.Background(), deadline)
	}

	return context.WithCancel(context.Background())
}

func PrivateTemp(t testing.TB) string {
	dst := t.TempDir()
	t.Setenv("TMPDIR", dst)
	return dst
}
