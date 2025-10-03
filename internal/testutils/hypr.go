package testutils

import (
	"bufio"
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func SetupHyprEnvVars(t *testing.T) (string, string) {
	tempDir := t.TempDir()
	signature := "test_signature"
	hyprDir := filepath.Join(tempDir, "hypr", signature)
	//nolint:gosec
	err := os.MkdirAll(hyprDir, 0o755)
	require.NoError(t, err, "failed to create hypr directory")

	originalXDG := os.Getenv("XDG_RUNTIME_DIR")
	originalSig := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	t.Cleanup(func() {
		_ = os.Setenv("XDG_RUNTIME_DIR", originalXDG)
		_ = os.Setenv("HYPRLAND_INSTANCE_SIGNATURE", originalSig)
	})

	err = os.Setenv("XDG_RUNTIME_DIR", tempDir)
	require.NoError(t, err, "failed to set env var")
	err = os.Setenv("HYPRLAND_INSTANCE_SIGNATURE", signature)
	require.NoError(t, err, "failed to set env var")
	return tempDir, signature
}

func SetupHyprSocket(ctx context.Context, t *testing.T, xdgRuntimeDir, signature string,
	hyprSocketFun func(string, string) string,
) (net.Listener, func()) {
	socketPath := hyprSocketFun(xdgRuntimeDir, signature)
	lc := &net.ListenConfig{}
	listener, err := lc.Listen(ctx, "unix", socketPath)
	require.NoError(t, err, "failed to create a test socket %s", socketPath)
	return listener, func() {
		err = listener.Close()
		require.NoError(t, err, "cant close hypr socket")
	}
}

func SetupFakeHyprIPCWriter(t *testing.T, listener net.Listener, responseData [][]byte,
	expectedCommands []string, exitOnError bool,
) chan struct{} {
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		Logf(t, "Starting hypr server")
		for i, command := range expectedCommands {
			Logf(t, "Sending message hypr server")
			respondToHyprIPC(t, listener, exitOnError, command, responseData, i)
		}
		Logf(t, "Ending hypr server")
	}()
	return serverDone
}

func respondToHyprIPC(t *testing.T, listener net.Listener, exitOnError bool, command string, responseData [][]byte, i int) {
	conn, err := listener.Accept()
	if err != nil && exitOnError {
		return
	}
	defer func() {
		if conn != nil {
			_ = conn.Close()
		}
	}()
	assert.Nil(t, err, "failed to accept connection")

	t.Log("listener accepted a connection")

	s := bufio.NewScanner(conn)
	if !s.Scan() {
		t.Errorf("no data to read")
	}
	buf := s.Bytes()
	assert.Nil(t, err, "failed to read")
	assert.Equal(t, command, string(buf), "wrong command")

	_, err = conn.Write(responseData[i])
	assert.Nil(t, err, "failed to write response")
	t.Logf("wrote response to the client %s", string(responseData[i]))
}

func SetupFakeHyprEventsServer(ctx context.Context, t *testing.T, listener net.Listener, events []string) chan struct{} {
	serverDone := make(chan struct{})
	go func() {
		defer func() {
			Logf(t, "Fake Hypr events server goroutine exiting")
			close(serverDone)
		}()

		Logf(t, "Fake Hypr events server: waiting for connection")
		conn, err := listener.Accept()
		if err != nil {
			t.Errorf("Failed to accept connection: %v", err)
			return
		}
		defer func() {
			Logf(t, "Fake Hypr events server: closing connection")
			if conn != nil {
				_ = conn.Close()
			}
		}()

		Logf(t, "Fake Hypr events server: accepted connection, ready to send events")

		for i, event := range events {
			select {
			case <-ctx.Done():
				Logf(t, "Fake Hypr events server: context cancelled while sending events")
				return
			default:
			}

			if _, err := conn.Write([]byte(event + "\n")); err != nil {
				t.Errorf("Failed to write event %d: %v", i, err)
				return
			}
			Logf(t, "Wrote event %d on the event socket: %s", i, event)
			time.Sleep(10 * time.Millisecond)
		}

		Logf(t, "Fake Hypr events server: finished sending %d events, waiting for context cancellation", len(events))
		<-ctx.Done()
		Logf(t, "Fake Hypr events server: context cancelled (%v), shutting down", context.Cause(ctx))
	}()
	return serverDone
}
