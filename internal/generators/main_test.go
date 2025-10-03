package generators_test

import (
	"flag"
	"os"
	"testing"
)

var regenerate = flag.Bool("regenerate", false, "regenerate fixtures instead of comparing")

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
