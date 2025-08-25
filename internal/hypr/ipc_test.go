package hypr_test

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestIPC_Run(t *testing.T) {
	tests := []struct {
		name           string
		mockEvents     []string
		instantlyClose bool
		expectedEvents []hypr.HyprEvent
		expectedErrors []string
		expectError    bool
		description    string
	}{
		{
			name: "happy_path",
			mockEvents: []string{
				"monitoraddedv2>>1,eDP-1,Built-in Display",
				"monitoraddedv2>>2,DP-1,External Monitor",
				"monitorremovedv2>>2,DP-1,External Monitor",
			},
			instantlyClose: false,
			expectedEvents: []hypr.HyprEvent{
				{
					Type: hypr.MonitorAdded,
					Monitor: &hypr.MonitorSpec{
						ID:          utils.IntPtr(1),
						Name:        "eDP-1",
						Description: "Built-in Display",
					},
				},
				{
					Type: hypr.MonitorAdded,
					Monitor: &hypr.MonitorSpec{
						ID:          utils.IntPtr(2),
						Name:        "DP-1",
						Description: "External Monitor",
					},
				},
				{
					Type: hypr.MonitorRemoved,
					Monitor: &hypr.MonitorSpec{
						ID:          utils.IntPtr(2),
						Name:        "DP-1",
						Description: "External Monitor",
					},
				},
			},
			expectError: false,
			description: "Should successfully process monitor add/remove events",
		},
		{
			name: "parse_error",
			mockEvents: []string{
				"monitoraddedv2>>placeholder",
			},
			instantlyClose: false,
			expectedErrors: []string{
				"cant parse event",
			},
			expectError: true,
			description: "Should return parse error for malformed event",
		},
		{
			name: "scanner_error",
			mockEvents: []string{
				"monitoraddedv2>>1,eDP-1,Built-in Display",
			},
			instantlyClose: false,
			expectedErrors: []string{
				"scanner error",
				"context canceled",
				"context deadline exceeded",
				"use of closed network connection",
				"operation was canceled",
			},
			expectError: true,
			description: "Should handle connection closed abruptly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown, listener, ipc := setupTest(t)
			defer teardown(t)

			expectedEventCount := len(tt.mockEvents)
			ipcDone, events := processIPCEvents(t, listener, tt.mockEvents, ipc, expectedEventCount, tt.instantlyClose)

			select {
			case err := <-ipcDone:
				if tt.expectError {
					if err != nil {
						if len(tt.expectedErrors) > 0 {
							errorFound := slices.ContainsFunc(tt.expectedErrors, func(expectedErr string) bool {
								return strings.Contains(err.Error(), expectedErr)
							})
							assert.True(t, errorFound, "Expected one of %v, got: %v", tt.expectedErrors, err)
						}
					} else {
						t.Errorf("Expected error but got nil")
					}
				} else {
					if err != nil && !strings.Contains(err.Error(), "context canceled") {
						t.Errorf("IPC returned unexpected error: %v", err)
					}
				}
			case <-time.After(1 * time.Second):
				t.Error("IPC didn't finish in time")
			}

			if !tt.expectError && len(tt.expectedEvents) > 0 {
				assert.Equal(t, tt.expectedEvents, events, tt.description)
			}
		})
	}
}

func processIPCEvents(t *testing.T, listener net.Listener, mockEvents []string, ipc *hypr.IPC, expectedEventCount int, instantlyClose bool) (chan error, []hypr.HyprEvent) {
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		conn, err := listener.Accept()
		if err != nil {
			t.Errorf("Failed to accept connection: %v", err)
			return
		}
		if !instantlyClose {
			defer conn.Close()
		} else {
			_ = conn.Close()
			return
		}

		for _, event := range mockEvents {
			if _, err := conn.Write([]byte(event + "\n")); err != nil {
				t.Errorf("Failed to write event: %v", err)
				return
			}
			time.Sleep(10 * time.Millisecond) // Small delay between events
		}

		// Wait a bit to ensure all events are processed
		time.Sleep(100 * time.Millisecond)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ipcDone := make(chan error, 1)
	go func() {
		ipcDone <- ipc.RunEventLoop(ctx)
	}()

	events := []hypr.HyprEvent{}
	eventsChannel := ipc.ListenEvents()

	eventTimeout := time.After(2 * time.Second)

collectLoop:
	for len(events) < expectedEventCount {
		select {
		case event, ok := <-eventsChannel:
			if !ok {
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
		{
			name:    "both_present",
			xdgVar:  "/tmp",
			sigVar:  "test_sig",
			wantErr: "",
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

			// Test NewIPC
			ipc, err := hypr.NewIPC()

			if tt.wantErr == "" {
				// Expect success
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

func setupTest(t *testing.T) (func(t *testing.T), net.Listener, *hypr.IPC) {
	tempDir := t.TempDir()
	hyprDir := filepath.Join(tempDir, "hypr", "test_signature")
	//nolint:gosec
	if err := os.MkdirAll(hyprDir, 0o755); err != nil {
		t.Fatalf("Failed to create hypr directory: %v", err)
	}

	originalXDG := os.Getenv("XDG_RUNTIME_DIR")
	originalSig := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	t.Cleanup(func() {
		_ = os.Setenv("XDG_RUNTIME_DIR", originalXDG)
		_ = os.Setenv("HYPRLAND_INSTANCE_SIGNATURE", originalSig)
	})

	_ = os.Setenv("XDG_RUNTIME_DIR", tempDir)
	_ = os.Setenv("HYPRLAND_INSTANCE_SIGNATURE", "test_signature")

	socketPath := hypr.GetHyprEventsSocket(tempDir, "test_signature")
	//nolint:noctx
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to create test socket: %v", err)
	}

	ipc, err := hypr.NewIPC()
	if err != nil {
		t.Fatalf("Failed to create IPC: %v", err)
	}

	teardown := func(t *testing.T) {
		_ = listener.Close()
	}

	return teardown, listener, ipc
}

func TestIPC_QueryConnectedMonitors(t *testing.T) {
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
			expectedError: "monitor spec is invalid",
			description:   "Should fail when monitor block is missing description",
		},
		{
			name:          "malformed_header",
			responseFile:  "testdata/monitors_response_malformed_header.json",
			expectedError: "monitor spec is invalid",
			description:   "Should fail when monitor header has wrong number of parts",
		},
		{
			name:          "malformed_description",
			responseFile:  "testdata/monitors_response_malformed_description.json",
			expectedError: "monitor spec is invalid",
			description:   "Should fail when description line has wrong format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			hyprDir := filepath.Join(tempDir, "hypr", "test_signature")
			//nolint:gosec
			if err := os.MkdirAll(hyprDir, 0o755); err != nil {
				t.Fatalf("Failed to create hypr directory: %v", err)
			}

			originalXDG := os.Getenv("XDG_RUNTIME_DIR")
			originalSig := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
			t.Cleanup(func() {
				_ = os.Setenv("XDG_RUNTIME_DIR", originalXDG)
				_ = os.Setenv("HYPRLAND_INSTANCE_SIGNATURE", originalSig)
			})

			_ = os.Setenv("XDG_RUNTIME_DIR", tempDir)
			_ = os.Setenv("HYPRLAND_INSTANCE_SIGNATURE", "test_signature")

			commandSocketPath := hypr.GetHyprSocket(tempDir, "test_signature")
			//nolint:noctx
			listener, err := net.Listen("unix", commandSocketPath)
			if err != nil {
				t.Fatalf("Failed to create test command socket: %v", err)
			}
			defer listener.Close()

			responseData, err := os.ReadFile(tt.responseFile)
			if err != nil {
				t.Fatalf("Failed to read test response file %s: %v", tt.responseFile, err)
			}

			serverDone := make(chan struct{})
			go func() {
				defer close(serverDone)
				conn, err := listener.Accept()
				if err != nil {
					t.Errorf("Failed to accept connection: %v", err)
					return
				}
				defer conn.Close()

				buf := make([]byte, 1024)
				_, err = conn.Read(buf)
				if err != nil {
					t.Errorf("Failed to read command: %v", err)
					return
				}

				_, err = conn.Write(responseData)
				if err != nil {
					t.Errorf("Failed to write response: %v", err)
					return
				}
			}()

			ipc, err := hypr.NewIPC()
			if err != nil {
				t.Fatalf("Failed to create IPC: %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			monitors, err := ipc.QueryConnectedMonitors(ctx)

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
			} else {
				assert.NoError(t, err, "Should not error for %s", tt.description)
				assert.Equal(t, tt.expectedMonitors, monitors, tt.description)
			}
		})
	}
}
