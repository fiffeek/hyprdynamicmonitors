package hypr_test

import (
	"flag"
	"os"
	"testing"
)

var debug = flag.Bool("debug", false, "enable debug logs")

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
