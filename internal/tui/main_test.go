package tui_test

import (
	"flag"
	"os"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var regenerate = flag.Bool("regenerate", false, "regenerate fixtures instead of comparing")

func TestMain(m *testing.M) {
	lipgloss.SetColorProfile(termenv.Ascii)
	os.Exit(m.Run())
}
