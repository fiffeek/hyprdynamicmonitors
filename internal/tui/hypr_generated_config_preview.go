package tui

import (
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/sirupsen/logrus"
)

type HDMGeneratedConfigPreview struct {
	cfg               *config.Config
	textarea          textarea.Model
	height            int
	width             int
	runningUnderTest  bool
	reloaded          bool
	destinationExists bool
}

func NewHDMGeneratedConfigPreview(cfg *config.Config,
	runningUnderTest bool,
) *HDMGeneratedConfigPreview {
	ta := textarea.New()
	return &HDMGeneratedConfigPreview{
		cfg:              cfg,
		textarea:         ta,
		runningUnderTest: runningUnderTest,
		reloaded:         false,
	}
}

func (h *HDMGeneratedConfigPreview) SetHeight(height int) {
	h.height = height
	h.textarea.SetHeight(height)
}

func (h *HDMGeneratedConfigPreview) SetWidth(width int) {
	h.width = width
	h.textarea.SetWidth(width)
}

func (h *HDMGeneratedConfigPreview) Clear() {
	h.reloaded = false
	h.destinationExists = false
}

func (h *HDMGeneratedConfigPreview) Update(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}

	switch msg := msg.(type) {
	case ConfigReloaded:
		logrus.Debug("Received config reloaded event")
		h.Clear()

	case OperationStatus:
		if msg.name == OperationNameHydrate || msg.name == OperationNameMatchingProfile {
			h.Clear()
		}
	// poll for the destination changes since the TUI does not watch for these
	case reloadHyprConfigMsg:
		logrus.Debug("Reload hypr config requested")
		h.Clear()
	}

	if !h.reloaded {
		h.reloaded = true

		contents, err := os.ReadFile(*h.cfg.Get().General.Destination)
		text := ""
		if err == nil {
			h.destinationExists = true
			text = string(contents)
		}

		if err == nil && text != h.textarea.Value() {
			cmds = append(cmds, OperationStatusCmd(OperationNameReloadHyprDestination, nil))
		}

		logrus.Debugf("Textarea text: %s", text)
		h.textarea.SetValue(text)

		logrus.Debug("Requesting reloading hypr config")
		cmds = append(cmds, reloadHyprConfig(5*time.Second))
	}

	return tea.Batch(cmds...)
}

func (h *HDMGeneratedConfigPreview) View() string {
	sections := []string{}
	availableHeight := h.height

	title := TitleStyle.Margin(0, 0, 0, 0).Render("Generated hypr config preview")
	availableHeight -= lipgloss.Height(title)
	sections = append(sections, title)

	configFile := *h.cfg.Get().General.Destination
	if h.runningUnderTest {
		configFile = filepath.Base(configFile)
	}

	subtitle := SubtitleInfoStyle.Margin(0, 0, 1, 0).Render(configFile)
	availableHeight -= lipgloss.Height(subtitle)
	sections = append(sections, subtitle)

	if h.destinationExists {
		h.textarea.SetWidth(h.width)
		h.textarea.SetHeight(availableHeight)
		ta := h.textarea.View()
		sections = append(sections, ta)
	} else {
		warnSections := []string{}
		warnSections = append(warnSections, "Are you running the daemon?")
		warnSections = append(warnSections, lipgloss.NewStyle().Width(h.width).Margin(0, 0,
			1, 0).Render("See: "+LinkStyle.Render("https://hyprdynamicmonitors.filipmikina.com/docs/quickstart/setup-approaches")))
		warnSections = append(warnSections, SubtitleInfoStyle.Width(h.width).Render(
			"Alternatively, run the generation once: `hyprdynamicmonitors run --run-once`"))
		sections = append(sections, lipgloss.JoinVertical(lipgloss.Top, warnSections...))
	}

	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}
