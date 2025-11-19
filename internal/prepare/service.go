// Package prepare defines utility service to prepare prior to hyprdynamicmonitors daemon run
package prepare

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/sirupsen/logrus"
)

type Service struct {
	cfg                 *config.Config
	monitorDisableRegex *regexp.Regexp
}

func NewService(cfg *config.Config) *Service {
	monitorDisableRegex := regexp.MustCompile(`.*monitor.*=.*disable.*`)
	return &Service{
		cfg,
		monitorDisableRegex,
	}
}

func (s *Service) TruncateDestination() error {
	file := s.cfg.Get().General.Destination
	_, err := os.Stat(*file)
	if err != nil {
		logrus.WithFields(logrus.Fields{"destination": *file}).Info("file does not exist")
		//nolint:nilerr
		return nil
	}

	contents, err := os.ReadFile(*file)
	if err != nil {
		return fmt.Errorf("cant read the %s destination file: %w", *file, err)
	}

	lines := strings.Split(string(contents), "\n")

	var filteredLines []string
	for _, line := range lines {
		if s.monitorDisableRegex.MatchString(line) {
			logrus.Infof("Line %s will be removed", line)
			continue
		}
		filteredLines = append(filteredLines, line)
	}

	newContents := strings.Join(filteredLines, "\n")
	if err := utils.WriteAtomic(*file, []byte(newContents)); err != nil {
		return fmt.Errorf("cant write to %s destination file: %w", *file, err)
	}

	return nil
}
