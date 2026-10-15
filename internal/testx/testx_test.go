package testx_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/egdaemon/fron/internal/testx"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	testx.Logging()
	os.Exit(m.Run())
}

func TestFixture(t *testing.T) {
	t.Run("should join paths under the fixtures directory", func(t *testing.T) {
		require.Equal(t, filepath.Join(".fixtures", "a", "b.txt"), testx.Fixture("a", "b.txt"))
	})
}

func TestRead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "content.txt")
	require.NoError(t, os.WriteFile(path, []byte("hello"), 0600))

	t.Run("should read a file as a reader", func(t *testing.T) {
		require.Equal(t, "hello", testx.IOString(testx.Read(path)))
	})

	t.Run("should read a file as a string", func(t *testing.T) {
		require.Equal(t, "hello", testx.ReadString(path))
	})

	t.Run("should read a reader as bytes", func(t *testing.T) {
		require.Equal(t, []byte("hello"), testx.IOBytes(bytes.NewBufferString("hello")))
	})

	t.Run("should panic when the file is missing", func(t *testing.T) {
		require.Panics(t, func() { testx.ReadString(filepath.Join(t.TempDir(), "missing")) })
	})
}

func TestMust(t *testing.T) {
	t.Run("should return the value", func(t *testing.T) {
		require.Equal(t, 1, testx.Must(1, nil))
	})

	t.Run("should panic on error", func(t *testing.T) {
		require.Panics(t, func() { testx.Must(1, errors.New("failed")) })
	})
}

func TestMustT(t *testing.T) {
	t.Run("should return the value when there is no error", func(t *testing.T) {
		require.Equal(t, "ok", testx.MustT("ok", nil)(t))
	})
}

func TestTempenvvar(t *testing.T) {
	t.Run("should set then restore the variable", func(t *testing.T) {
		t.Setenv("FRON_TESTX_VAR", "before")
		testx.Tempenvvar("FRON_TESTX_VAR", "during", func() {
			require.Equal(t, "during", os.Getenv("FRON_TESTX_VAR"))
		})
		require.Equal(t, "before", os.Getenv("FRON_TESTX_VAR"))
	})
}

func TestContext(t *testing.T) {
	t.Run("should be cancelable", func(t *testing.T) {
		ctx, done := testx.Context(t)
		done()
		require.Error(t, ctx.Err())
	})
}

func TestPrivateTemp(t *testing.T) {
	t.Run("should point TMPDIR at a private directory", func(t *testing.T) {
		dir := testx.PrivateTemp(t)
		require.Equal(t, dir, os.Getenv("TMPDIR"))
		require.DirExists(t, dir)
	})
}
