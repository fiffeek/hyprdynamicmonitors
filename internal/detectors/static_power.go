package detectors

import (
	"context"
	"fmt"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type StaticPowerDetector struct {
	cfg    *config.PowerSection
	events chan PowerEvent
}

func NewStaticPowerDetector(cfg *config.PowerSection) *StaticPowerDetector {
	logrus.Debug("Running static power detector, no events will be sent or monitored")
	return &StaticPowerDetector{
		cfg:    cfg,
		events: make(chan PowerEvent, 10),
	}
}

func (p *StaticPowerDetector) GetCurrentState(ctx context.Context) (PowerState, error) {
	return ACPower, nil
}

func (p *StaticPowerDetector) Listen() <-chan PowerEvent {
	return p.events
}

func (p *StaticPowerDetector) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()
		logrus.Debug("Power detector context cancelled, closing D-Bus connection")
		return ctx.Err()
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("goroutines for static power detector failed %w", err)
	}
	return nil
}
