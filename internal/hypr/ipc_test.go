package hypr_test

import (
	"context"
	"net"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/testutils"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIPC_Run(t *testing.T) {
	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	tests := []struct {
		name               string
		mockEvents         []string
		responsePaths      []string
		expectedEventPaths []string
		expectedErrors     []string
		expectedCommands   []string
		expectError        bool
		description        string
		exitOnError        bool
	}{
		{
			name: "happy_path",
			mockEvents: []string{
				"monitoraddedv2>>1,eDP-1,Built-in Display",
				"monitoraddedv2>>2,DP-1,External Monitor",
				"monitorremovedv2>>2,DP-1,External Monitor",
			},
			expectedCommands: []string{
				"j/monitors all", // initial setup
				"j/monitors all", // first event
				"j/monitors all",
				"j/monitors all",
			},
			responsePaths: []string{
				"testdata/monitors_response_valid_1.json", // initial setup
				"testdata/monitors_response_valid_2.json", // first event
				"testdata/monitors_response_valid_1.json",
				"testdata/monitors_response_valid_2.json",
			},
			expectedEventPaths: []string{
				"testdata/monitors_response_valid_2.json",
				"testdata/monitors_response_valid_1.json",
				"testdata/monitors_response_valid_2.json",
			},
			expectError: false,
			description: "Should successfully process monitor add/remove events",
		},
		{
			name: "happy_path_dedup",
			mockEvents: []string{
				"monitoraddedv2>>1,eDP-1,Built-in Display",
				"monitoraddedv2>>2,DP-1,External Monitor",
				"monitoraddedv2>>2,DP-1,External Monitor",
			},
			expectedCommands: []string{
				"j/monitors all", // initial setup
				"j/monitors all", // first event
				"j/monitors all",
				"j/monitors all",
			},
			responsePaths: []string{
				"testdata/monitors_response_valid_1.json", // initial setup
				"testdata/monitors_response_valid_1.json", // first event
				"testdata/monitors_response_valid_1.json", // second event
				"testdata/monitors_response_valid_2.json", // third event
			},
			expectedEventPaths: []string{
				"testdata/monitors_response_valid_1.json", // first event is always sent
				"testdata/monitors_response_valid_2.json", // third event
			},
			expectError: false,
			description: "Should not send events when data does not change",
		},
		{
			name: "parse_error",
			mockEvents: []string{
				"monitoraddedv2>>placeholder",
			},
			expectedErrors: []string{
				"cant parse event",
			},
			responsePaths: []string{
				"testdata/monitors_response_valid_1.json",
			},
			expectedCommands: []string{
				"j/monitors all",
			},
			expectedEventPaths: []string{
				"testdata/monitors_response_valid_1.json",
			},
			expectError: true,
			description: "Should return parse error for malformed event",
			exitOnError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			responseData := [][]byte{}
			for _, path := range tt.responsePaths {
				// nolint:gosec
				data, err := os.ReadFile(path)
				assert.Nil(t, err, "Failed to read test response file %s: %w", path, err)
				responseData = append(responseData, data)
				res := &hypr.MonitorSpecs{}
				assert.Nil(t, utils.UnmarshalResponse(data, &res), "cant parse %s", path)
			}

			expectedEvents := []hypr.MonitorSpecs{}
			for _, path := range tt.expectedEventPaths {
				// nolint:gosec
				data, err := os.ReadFile(path)
				assert.Nil(t, err, "Failed to read test expected event file %s: %w", path, err)
				res := hypr.MonitorSpecs{}
				assert.Nil(t, utils.UnmarshalResponse(data, &res), "cant parse %s", path)
				expectedEvents = append(expectedEvents, res)
			}

			xdgRuntimeDir, signature := testutils.SetupHyprEnvVars(t)
			eventsListener, eventsSocketCleanUp := testutils.SetupHyprSocket(
				ctx,
				t,
				xdgRuntimeDir,
				signature,
				hypr.GetHyprEventsSocket)
			ipcListener, ipcCleanUp := testutils.SetupHyprSocket(ctx, t, xdgRuntimeDir, signature, hypr.GetHyprSocket)
			writerDone := testutils.SetupFakeHyprIPCWriter(t, ipcListener, responseData, tt.expectedCommands, tt.exitOnError)
			ipc, err := hypr.NewIPC(ctx)
			require.NoError(t, err, "failed to create ipc")
			defer func() {
				eventsSocketCleanUp()
				ipcCleanUp()
			}()

			expectedEventCount := len(expectedEvents)
			ipcDone, events := processIPCEvents(t, ctx, cancel, eventsListener, tt.mockEvents, ipc, expectedEventCount)

			select {
			case <-writerDone:
				break
			case err := <-ipcDone:
				if !tt.expectError && expectedEventCount > 0 {
					assert.Equal(t, expectedEvents, events, tt.description)
					return
				}
				if !tt.expectError && err != nil && !strings.Contains(err.Error(), "context canceled") {
					t.Errorf("IPC returned unexpected error: %v", err)
					return
				}
				if tt.expectError && err != nil {
					if len(tt.expectedErrors) > 0 {
						errorFound := slices.ContainsFunc(tt.expectedErrors, func(expectedErr string) bool {
							return strings.Contains(err.Error(), expectedErr)
						})
						assert.True(t, errorFound, "Expected one of %v, got: %v", tt.expectedErrors, err)
						return
					}
				}
				if tt.expectError && err == nil {
					t.Errorf("Expected error but got nil")
					return
				}
			case <-time.After(1 * time.Second):
				t.Error("IPC didn't finish in time")
			}
		})
	}
}

func processIPCEvents(t *testing.T, ctx context.Context, cancel context.CancelFunc, listener net.Listener, mockEvents []string, ipc *hypr.IPC,
	expectedEventCount int,
) (chan error, []hypr.MonitorSpecs) {
	serverDone := testutils.SetupFakeHyprEventsServer(t, listener, mockEvents)

	ipcDone := make(chan error, 1)
	go func() {
		ipcDone <- ipc.RunEventLoop(ctx)
	}()

	events := []hypr.MonitorSpecs{}
	eventsChannel := ipc.Listen()
	eventTimeout := time.After(2 * time.Second)

	t.Log("collecting events")
collectLoop:
	for len(events) < expectedEventCount {
		select {
		case event, ok := <-eventsChannel:
			if !ok {
				t.Log("channel for events is closed")
				break collectLoop
			}
			events = append(events, event)
		case <-eventTimeout:
			t.Errorf("Timeout waiting for events. Got %d events, expected %d", len(events), expectedEventCount)
			break collectLoop
		}
	}

	cancel()
	select {
	case <-serverDone:
	case <-time.After(1 * time.Second):
		t.Error("Server didn't finish in time")
	}

	return ipcDone, events
}

func TestNewIPC_MissingEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name    string
		xdgVar  string
		sigVar  string
		wantErr string
	}{
		{
			name:    "missing_hyprland_signature",
			xdgVar:  "/tmp",
			sigVar:  "",
			wantErr: "HYPRLAND_INSTANCE_SIGNATURE environment variable not set",
		},
		{
			name:    "missing_xdg_runtime_dir",
			xdgVar:  "",
			sigVar:  "test_sig",
			wantErr: "XDG_RUNTIME_DIR environment variable not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			originalXDG := os.Getenv("XDG_RUNTIME_DIR")
			originalSig := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")

			t.Cleanup(func() {
				_ = os.Setenv("XDG_RUNTIME_DIR", originalXDG)
				_ = os.Setenv("HYPRLAND_INSTANCE_SIGNATURE", originalSig)
			})

			if tt.xdgVar == "" {
				os.Unsetenv("XDG_RUNTIME_DIR")
			} else {
				_ = os.Setenv("XDG_RUNTIME_DIR", tt.xdgVar)
			}

			if tt.sigVar == "" {
				os.Unsetenv("HYPRLAND_INSTANCE_SIGNATURE")
			} else {
				_ = os.Setenv("HYPRLAND_INSTANCE_SIGNATURE", tt.sigVar)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			ipc, err := hypr.NewIPC(ctx)

			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if ipc == nil {
					t.Error("Expected IPC instance, got nil")
				}
			} else {
				// Expect error
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Expected error containing '%s', got: %v", tt.wantErr, err)
				}
				if ipc != nil {
					t.Error("Expected nil IPC instance on error")
				}
			}
		})
	}
}

func TestIPC_GetConnectedMonitors(t *testing.T) {
	tests := []struct {
		name             string
		responseFile     string
		expectedMonitors hypr.MonitorSpecs
		expectedError    string
		description      string
	}{
		{
			name:         "happy_path",
			responseFile: "testdata/monitors_response_valid.json",
			expectedMonitors: []*hypr.MonitorSpec{
				{
					Name:        "eDP-1",
					ID:          utils.IntPtr(0),
					Description: "BOE NE135A1M-NY1",
				},
				{
					Name:        "DP-11",
					ID:          utils.IntPtr(1),
					Description: "LG Electronics LG SDQHD 301NTBKDU037",
				},
			},
			description: "Should successfully parse valid monitor response",
		},
		{
			name:          "missing_description",
			responseFile:  "testdata/monitors_response_missing_description.json",
			expectedError: "failed to validate response: invalid monitor: desc cant be empty",
			description:   "Should fail when monitor block is missing description",
		},
		{
			name:          "malformed_header",
			responseFile:  "testdata/monitors_response_malformed_header.json",
			expectedError: "failed to validate response: invalid monitor: id cant be nil",
			description:   "Should fail when monitor header has wrong number of parts",
		},
		{
			name:          "malformed_description",
			responseFile:  "testdata/monitors_response_malformed_description.json",
			expectedError: "failed to validate response: invalid monitor: desc cant be empty",
			description:   "Should fail when description line has wrong format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			xdgRuntimeDir, signature := testutils.SetupHyprEnvVars(t)
			listener, teardown := testutils.SetupHyprSocket(ctx, t, xdgRuntimeDir, signature, hypr.GetHyprSocket)
			defer teardown()

			responseData, err := os.ReadFile(tt.responseFile)
			assert.Nil(t, err, "Failed to read test response file %s: %w", tt.responseFile, err)

			serverDone := testutils.SetupFakeHyprIPCWriter(t, listener, [][]byte{
				responseData,
			}, []string{"j/monitors all"}, false)

			ipc, err := hypr.NewIPC(ctx)
			var monitors hypr.MonitorSpecs
			if ipc != nil {
				monitors = ipc.GetConnectedMonitors()
			}

			select {
			case <-serverDone:
			case <-time.After(1 * time.Second):
				t.Error("Server didn't finish in time")
			}

			if tt.expectedError != "" {
				assert.Error(t, err, "Expected error for %s", tt.description)
				if err != nil {
					assert.Contains(t, err.Error(), tt.expectedError, "Error should contain expected text")
				}
				return
			}

			assert.NoError(t, err, "Should not error for %s", tt.description)
			assert.Equal(t, tt.expectedMonitors, monitors, tt.description)
		})
	}
}
