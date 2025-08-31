// Package filewatcher provides a service that watches config files and issues
// a debounced event with changes
package filewatcher

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	cfg                  *config.Config
	stateMu              sync.RWMutex
	watchedPaths         map[string]bool
	watchingConfig       bool
	events               chan interface{}
	watcher              *fsnotify.Watcher
	disableAutoHotReload *bool
	debouncer            *utils.Debouncer
}

func NewService(cfg *config.Config, disableAutoHotReload *bool) *Service {
	return &Service{
		cfg: cfg, stateMu: sync.RWMutex{}, watchedPaths: make(map[string]bool),
		events: make(chan interface{}, 1), watchingConfig: false,
		disableAutoHotReload: disableAutoHotReload,
		debouncer:            utils.NewDebouncer(),
	}
}

func (s *Service) Update() error {
	logrus.Debug("Updating watcher")

	if s.disableAutoHotReload != nil && *s.disableAutoHotReload {
		logrus.Info("Hot reload disabled, not updating filewatcher")
		return nil
	}

	if s.watcher == nil {
		return errors.New("no watcher assigned")
	}

	s.stateMu.Lock()
	defer s.stateMu.Unlock()

	if err := s.ensureConfigPathIsTracked(); err != nil {
		return fmt.Errorf("cant track config: %w", err)
	}

	if s.checkIfAllPathsAreAlreadyTracked() {
		logrus.Info("All paths are tracked already, no update needed")
		return nil
	}

	if err := s.removeCurrentlyTrackedPaths(); err != nil {
		return fmt.Errorf("cant remove currently tracked paths: %w", err)
	}

	if err := s.addWatchedPaths(); err != nil {
		return fmt.Errorf("cant add new tracked paths: %w", err)
	}

	logrus.Debug("Watcher update done")

	return nil
}

func (s *Service) addWatchedPaths() error {
	logrus.Debug("(Re)adding new tracked paths")
	for _, profile := range s.cfg.Get().Profiles {
		if _, ok := s.watchedPaths[profile.ConfigFileDir]; ok {
			continue
		}
		if err := s.watcher.Add(profile.ConfigFileDir); err != nil {
			return fmt.Errorf("cant add %s to watcher: %w", profile.ConfigFileDir, err)
		}
		s.watchedPaths[profile.ConfigFileDir] = true
		logrus.WithFields(logrus.Fields{"path": profile.ConfigFileDir}).Debug("Added watched path")
	}
	return nil
}

func (s *Service) removeCurrentlyTrackedPaths() error {
	logrus.Debug("Removing currently tracked paths")
	for path := range s.watchedPaths {
		if _, ok := s.watchedPaths[path]; !ok {
			continue
		}
		if err := s.watcher.Remove(path); err != nil &&
			errors.Is(err, fsnotify.ErrNonExistentWatch) {
			return fmt.Errorf("cant remove %s from watcher: %w", path, err)
		}
		delete(s.watchedPaths, path)
		logrus.WithFields(logrus.Fields{"path": path}).Debug("Removed watched path")
	}
	return nil
}

func (s *Service) checkIfAllPathsAreAlreadyTracked() bool {
	allPathsTracked := true
	for _, profile := range s.cfg.Get().Profiles {
		if _, ok := s.watchedPaths[profile.ConfigFileDir]; !ok {
			allPathsTracked = false
			break
		}
	}
	return allPathsTracked
}

func (s *Service) ensureConfigPathIsTracked() error {
	if s.watchingConfig {
		return nil
	}
	if err := s.watcher.Add(s.cfg.Get().ConfigDirPath); err != nil {
		return fmt.Errorf("cant watch config dir %s: %w", s.cfg.Get().ConfigDirPath, err)
	}
	s.watchingConfig = true
	logrus.WithFields(logrus.Fields{
		"config_dir": s.cfg.Get().ConfigDirPath,
		"config":     s.cfg.Get().ConfigPath,
	}).Debug("Added config path to watchlist")
	return nil
}

func (s *Service) Listen() <-chan interface{} {
	return s.events
}

func (s *Service) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		<-ctx.Done()
		logrus.Debug("Context cancelled for filewatcher, shutting down")
		return context.Cause(ctx)
	})

	if s.disableAutoHotReload != nil && *s.disableAutoHotReload {
		logrus.Info("Disabling filewatcher")
		if err := eg.Wait(); err != nil {
			return fmt.Errorf("bg tasks failed in filewatcher: %w", err)
		}
		return nil
	}

	logrus.Debug("starting watcher")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("cant create watcher: %w", err)
	}
	s.watcher = watcher

	eg.Go(func() error {
		return s.debouncer.Run(ctx)
	})

	eg.Go(func() error {
		<-ctx.Done()
		s.debouncer.Cancel()
		logrus.Debug("Context cancelled, shutting watcher down")
		if err := watcher.Close(); err != nil {
			logrus.WithError(err).Error("Cant close watcher on exit")
		}
		return context.Cause(ctx)
	})

	eg.Go(func() error {
		logrus.Debug("Initialized watcher")
		if err := s.runServiceLoop(ctx, watcher); err != nil {
			return fmt.Errorf("cant run service loop: %w", err)
		}
		logrus.Debug("Exiting watcher")
		return nil
	})

	if err := s.Update(); err != nil {
		return fmt.Errorf("cant initialize watcher: %w", err)
	}

	return eg.Wait()
}

func (s *Service) runServiceLoop(ctx context.Context, watcher *fsnotify.Watcher) error {
	logrus.Debug("Starting filewatcher goroutine")
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return errors.New("watcher channel is closed")
			}

			logrus.WithFields(logrus.Fields{
				"name":      event.Name,
				"operation": event.Op,
			}).Debug("Received filewatcher event")

			s.debouncer.Do(ctx, time.Duration(*s.cfg.Get().HotReload.UpdateDebounceTimer)*time.Millisecond, s.updateProcessor)
			logrus.WithFields(logrus.Fields{"fun": s.updateProcessor}).Debug("Scheduled debounced update")
		case err, ok := <-watcher.Errors:
			if !ok {
				return errors.New("watcher error channel is closed")
			}
			if err != nil {
				return fmt.Errorf("watcher error received: %w", err)
			}
		case <-ctx.Done():
			logrus.Debug("Context cancelled, shutting fswatcher down")
			return context.Cause(ctx)
		}
	}
}

func (s *Service) updateProcessor(ctx context.Context) error {
	select {
	case <-ctx.Done():
		logrus.Debug("Config update processor context cancelled, shutting down")
		return context.Cause(ctx)
	default:
		logrus.Debug("Sending update event")
		s.events <- true
		return nil
	}
}
