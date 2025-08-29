package reloader_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fiffeek/hyprdynamicmonitors/internal/reloader"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePowerDetector struct {
	reloadErr   error
	reloadCalls int
}

func (f *fakePowerDetector) Reload(ctx context.Context) error {
	f.reloadCalls++
	return f.reloadErr
}

type fakeService struct {
	updateErr   error
	updateCalls int
}

func (f *fakeService) UpdateOnce() error {
	f.updateCalls++
	return f.updateErr
}

type fakeFilewatcher struct {
	updateErr   error
	updateCalls int
	channel     chan interface{}
}

func (f *fakeFilewatcher) Update() error {
	f.updateCalls++
	return f.updateErr
}

func (f *fakeFilewatcher) Listen() <-chan interface{} {
	return f.channel
}

func TestService_Reload(t *testing.T) {
	ctx := context.Background()
	cfg := testutils.NewTestConfig(t).Get()

	tests := []struct {
		name           string
		powerErr       error
		serviceErr     error
		filewatcherErr error
		wantErr        bool
		errContains    string
	}{
		{
			name:    "successful reload",
			wantErr: false,
		},
		{
			name:           "filewatcher update fails",
			filewatcherErr: errors.New("filewatcher error"),
			wantErr:        true,
			errContains:    "cant update filewatcher",
		},
		{
			name:        "power detector reload fails",
			powerErr:    errors.New("power detector error"),
			wantErr:     true,
			errContains: "cant reload powerDetector",
		},
		{
			name:        "service update fails",
			serviceErr:  errors.New("service error"),
			wantErr:     true,
			errContains: "cant update user service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			powerDetector := &fakePowerDetector{reloadErr: tt.powerErr}
			service := &fakeService{updateErr: tt.serviceErr}
			filewatcher := &fakeFilewatcher{updateErr: tt.filewatcherErr}

			reloaderService := reloader.NewService(cfg, filewatcher, powerDetector, service, false)

			err := reloaderService.Reload(ctx)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 1, filewatcher.updateCalls)
				assert.Equal(t, 1, powerDetector.reloadCalls)
				assert.Equal(t, 1, service.updateCalls)
			}
		})
	}
}

func TestService_Run(t *testing.T) {
	cfg := testutils.NewTestConfig(t).Get()

	tests := []struct {
		name                     string
		hotReloadDisabled        bool
		expectedFilewatcherCalls int
		expectedPowerCalls       int
		expectedServiceCalls     int
	}{
		{
			name:                     "processes events from filewatcher",
			hotReloadDisabled:        false,
			expectedFilewatcherCalls: 1,
			expectedPowerCalls:       1,
			expectedServiceCalls:     1,
		},
		{
			name:                     "disabled hot reload",
			hotReloadDisabled:        true,
			expectedFilewatcherCalls: 0,
			expectedPowerCalls:       0,
			expectedServiceCalls:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			powerDetector := &fakePowerDetector{}
			service := &fakeService{}
			channel := make(chan interface{}, 1)
			filewatcher := &fakeFilewatcher{channel: channel}

			reloaderService := reloader.NewService(cfg, filewatcher, powerDetector, service, tt.hotReloadDisabled)

			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			// Start the service in a goroutine
			errCh := make(chan error, 1)
			go func() {
				errCh <- reloaderService.Run(ctx)
			}()

			// Send an event
			channel <- true

			// Wait a bit to let them process
			time.Sleep(200 * time.Millisecond)

			cancel()

			select {
			case err := <-errCh:
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "context canceled")
			case <-time.After(100 * time.Millisecond):
				t.Fatal("timeout waiting for service to shutdown")
			}

			assert.Equal(t, tt.expectedFilewatcherCalls, filewatcher.updateCalls)
			assert.Equal(t, tt.expectedPowerCalls, powerDetector.reloadCalls)
			assert.Equal(t, tt.expectedServiceCalls, service.updateCalls)
		})
	}
}
