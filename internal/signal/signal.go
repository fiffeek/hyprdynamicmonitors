// Package signal provides signal handling functionality.
package signal

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Interrupted struct{ sig os.Signal }

func (i Interrupted) Error() string { return "interrupted by " + i.sig.String() }

func (i Interrupted) ExitCode() int {
	if s, ok := i.sig.(syscall.Signal); ok {
		return 128 + int(s)
	}
	return 1
}

type SIGUSR1Handler interface {
	Handle(context.Context) error
}

type SIGHUPHandler interface {
	Handle(context.Context) error
}

type Handler struct {
	sigChan     chan os.Signal
	cancelCause context.CancelCauseFunc
	sighup      SIGHUPHandler
	sigurs1     SIGUSR1Handler
}

func NewHandler(cancelCause context.CancelCauseFunc, sighup SIGHUPHandler, sigusr1 SIGUSR1Handler) *Handler {
	return &Handler{
		sigChan:     make(chan os.Signal, 1),
		cancelCause: cancelCause,
		sighup:      sighup,
		sigurs1:     sigusr1,
	}
}

func (h *Handler) Run(ctx context.Context) error {
	signal.Notify(h.sigChan, syscall.SIGUSR1, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	logrus.Debug("Signal notifications registered for SIGUSR1, SIGTERM, SIGINT, SIGHUP")

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()
		logrus.Debug("Signal handler goroutine exiting")
		signal.Stop(h.sigChan)
		return nil
	})

	eg.Go(func() error {
		logrus.Debug("Signal handler goroutine started")
		defer close(h.sigChan)
		for {
			select {
			case sig := <-h.sigChan:
				logrus.WithField("signal", sig).Debug("Signal received")
				switch sig {
				case syscall.SIGHUP:
					logrus.Info("Received SIGHUP")
					if err := h.sighup.Handle(ctx); err != nil {
						return fmt.Errorf("SIGUSR1 handler failed: %w", err)
					} else {
						logrus.Info("SIGHUP handled normally")
					}
				case syscall.SIGUSR1:
					logrus.Info("Received SIGUSR1")
					if err := h.sigurs1.Handle(ctx); err != nil {
						return fmt.Errorf("error while handling SIGUSR1: %w", err)
					} else {
						logrus.Info("SIGUSR1 handled normally")
					}
				case syscall.SIGTERM, syscall.SIGINT:
					logrus.WithField("signal", sig).Info("Received termination signal, shutting down gracefully")
					h.cancelCause(&Interrupted{sig})
					return &Interrupted{sig}
				}
			case <-ctx.Done():
				logrus.Debug("Signal handler context done, exiting")
				return context.Cause(ctx)
			}
		}
	})

	return eg.Wait()
}
