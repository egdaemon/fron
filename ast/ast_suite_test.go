package ast_test

import (
	"os"
	"testing"

	"github.com/egdaemon/fron/internal/testx"
)

func TestMain(m *testing.M) {
	testx.Logging()
	os.Exit(m.Run())
}
