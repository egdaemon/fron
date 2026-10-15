package ast_test

import (
	"testing"
	"time"

	"github.com/egdaemon/fron/ast"
	"github.com/egdaemon/fron/engine"
	"github.com/egdaemon/fron/internal/testx"
	"github.com/stretchr/testify/require"
)

func TestCompile(t *testing.T) {
	t.Run("should match the code first option form", func(t *testing.T) {
		a, err := ast.Parse("fron.zero.offset | 15m @ on weekdays during 09:00-17:00 in UTC")
		require.NoError(t, err)

		compiled, err := ast.Compile(a, "")
		require.NoError(t, err)

		manual, err := engine.New(
			15*time.Minute,
			engine.ZeroOffset,
			engine.Options().
				Weekdays().
				During(9*time.Hour, 17*time.Hour).
				InTz(time.UTC)...,
		)
		require.NoError(t, err)

		for _, from := range []time.Time{
			time.Date(2026, 10, 15, 8, 0, 0, 0, time.UTC),
			time.Date(2026, 10, 16, 16, 59, 0, 0, time.UTC),
			time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC),
		} {
			require.Equal(t, testx.MustT(manual.Next(from))(t), testx.MustT(compiled.Next(from))(t), from)
		}
	})

	t.Run("should use the seed for the offset", func(t *testing.T) {
		a, err := ast.Parse("billing-job | 2h")
		require.NoError(t, err)

		compiled, err := ast.Compile(a, "")
		require.NoError(t, err)

		manual, err := engine.New(2*time.Hour, engine.Seed("billing-job"))
		require.NoError(t, err)

		from := time.Date(2026, 10, 15, 8, 0, 0, 0, time.UTC)
		require.Equal(t, testx.MustT(manual.Next(from))(t), testx.MustT(compiled.Next(from))(t))
	})

	t.Run("should let extra options override the tree", func(t *testing.T) {
		a, err := ast.Parse("fron.zero.offset | 1h @ on weekdays")
		require.NoError(t, err)

		compiled, err := ast.Compile(a, "", engine.Options().Weekends()...)
		require.NoError(t, err)

		// 2026-10-16 is a friday.
		from := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
		require.Equal(t, time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC), testx.MustT(compiled.Next(from))(t))
	})
}
