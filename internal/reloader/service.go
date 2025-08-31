// Package reloader provides a service that listens to file change notifications
// and issues an application-wide reload
package reloader

import (
	"context"
	"errors"
	"fmt"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type IPowerDetector interface {
	Reload(context.Context) error
}

type IService interface {
	UpdateOnce(context.Context) error
}

type IFilewatcher interface {
	Update() error
	Listen() <-chan interface{}
}

type Service struct {
	cfg                  *config.Config
	filewatcher          IFilewatcher
	powerDetector        IPowerDetector
	service              IService
	disableAutoHotReload *bool
}

func NewService(cfg *config.Config, filewatcher IFilewatcher, powerDetector IPowerDetector,
	service IService, disableAutoHotReload bool,
) *Service {
	return &Service{
		cfg,
		filewatcher,
		powerDetector,
		service,
		&disableAutoHotReload,
	}
}

func (s *Service) Handle(ctx context.Context) error {
	return s.Reload(ctx)
}

func (s *Service) Reload(ctx context.Context) error {
	updates := []struct {
		Fun  func() error
		Name string
		Err  string
	}{
		{Fun: s.cfg.Reload, Name: "config reload", Err: "cant reload configuration"},
		{Fun: s.filewatcher.Update, Name: "update filewatcher", Err: "cant update filewatcher"},
		{Fun: func() error { return s.powerDetector.Reload(ctx) }, Name: "power detector reload", Err: "cant reload powerDetector"},
		{
			Fun:  func() error { return s.service.UpdateOnce(ctx) },
			Name: "updating user configuration", Err: "cant update user service",
		},
	}

	for _, update := range updates {
		logrus.Debug("Executing " + update.Name)
		if err := update.Fun(); err != nil {
			return fmt.Errorf("%s: %w", update.Err, err)
		}
	}

	return nil
}

func (s *Service) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		<-ctx.Done()
		logrus.Debug("Context cancelled for reloader, shutting down")
		return context.Cause(ctx)
	})

	if s.disableAutoHotReload != nil && *s.disableAutoHotReload {
		logrus.Info("Disabling reloader, no files will be watched")
		return eg.Wait()
	}

	watcherEventsChannel := s.filewatcher.Listen()

	eg.Go(func() error {
		logrus.Debug("Reloader event processor starting")
		for {
			select {
			case _, ok := <-watcherEventsChannel:
				if !ok {
					return errors.New("watcher event channel closed")
				}
				logrus.Debug("Watcher event received")
				if err := s.Reload(ctx); err != nil {
					return fmt.Errorf("cant reload user configuraton: %w", err)
				}

			case <-ctx.Done():
				logrus.Debug("Reloader event processor context cancelled, shutting down")
				return context.Cause(ctx)

			}
		}
	})

	return eg.Wait()
}
