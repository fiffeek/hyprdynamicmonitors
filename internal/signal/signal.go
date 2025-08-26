// Package signal provides signal handling functionality.
package signal

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

type Handler struct {
	sigChan     chan os.Signal
	ctx         context.Context
	cancelCause context.CancelCauseFunc
}

type SignalHandler interface {
	RunOnce(context.Context) error
}

func NewHandler(ctx context.Context, cancelCause context.CancelCauseFunc) *Handler {
	return &Handler{
		sigChan:     make(chan os.Signal, 1),
		ctx:         ctx,
		cancelCause: cancelCause,
	}
}

func (h *Handler) Start(handler SignalHandler) {
	signal.Notify(h.sigChan, syscall.SIGUSR1, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	logrus.Debug("Signal notifications registered for SIGUSR1, SIGTERM, SIGINT, SIGHUP")
	logrus.WithField("channel_cap", cap(h.sigChan)).Info("Signal channel capacity")

	go h.handleSignals(handler)
	logrus.Debug("Signal handler goroutine launched")
}

func (h *Handler) Stop() {
	h.cancelCause(context.Canceled)
	signal.Stop(h.sigChan)
	close(h.sigChan)
}

func (h *Handler) handleSignals(handler SignalHandler) {
	logrus.Debug("Signal handler goroutine started")
	for {
		select {
		case sig := <-h.sigChan:
			logrus.WithField("signal", sig).Debug("Signal received")
			switch sig {
			case syscall.SIGUSR1:
				logrus.Info("Received SIGUSR1, triggering manual update")
				if err := handler.RunOnce(h.ctx); err != nil {
					logrus.WithError(err).Error("Manual update failed, service will keep running")
				} else {
					logrus.Info("Manual update completed successfully")
				}
			case syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP:
				logrus.WithField("signal", sig).Info("Received termination signal, shutting down gracefully")
				h.cancelCause(context.Canceled)
				return
			}
		case <-h.ctx.Done():
			logrus.Debug("Signal handler context done, exiting")
			return
		}
	}
}
