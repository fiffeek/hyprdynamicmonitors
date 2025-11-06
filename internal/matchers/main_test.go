package matchers_test

import (
	"flag"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

var debug = flag.Bool("debug", false, "run with logrus debug")

func TestMain(m *testing.M) {
	flag.Parse()
	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	os.Exit(m.Run())
}
