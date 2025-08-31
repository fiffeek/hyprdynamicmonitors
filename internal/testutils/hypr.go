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
)

func SetupHyprEnvVars(t *testing.T) (string, string) {
	tempDir := t.TempDir()
	signature := "test_signature"
	hyprDir := filepath.Join(tempDir, "hypr", signature)
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
	_ = os.Setenv("HYPRLAND_INSTANCE_SIGNATURE", signature)
	return tempDir, signature
}

func SetupHyprSocket(ctx context.Context, t *testing.T, xdgRuntimeDir, signature string,
	hyprSocketFun func(string, string) string,
) (net.Listener, func()) {
	socketPath := hyprSocketFun(xdgRuntimeDir, signature)
	lc := &net.ListenConfig{}
	listener, err := lc.Listen(ctx, "unix", socketPath)
	assert.NoError(t, err, "failed to create a test socket %s", socketPath)
	return listener, func() { _ = listener.Close() }
}

func SetupFakeHyprIPCWriter(t *testing.T, listener net.Listener, responseData [][]byte,
	expectedCommands []string, exitOnError bool,
) chan struct{} {
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		for i, command := range expectedCommands {
			respondToHyprIPC(t, listener, exitOnError, command, responseData, i)
		}
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
	assert.Nil(t, err, "failded to read")
	assert.Equal(t, command, string(buf), "wrong command")

	_, err = conn.Write(responseData[i])
	assert.Nil(t, err, "failded to write response")
	t.Logf("wrote response to the client %s", string(responseData[i]))
}

func SetupFakeHyprEventsServer(t *testing.T, listener net.Listener, events []string) chan struct{} {
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		conn, err := listener.Accept()
		if err != nil {
			t.Errorf("Failed to accept connection: %v", err)
			return
		}

		t.Log("Accepted connection on events socket")

		for _, event := range events {
			if _, err := conn.Write([]byte(event + "\n")); err != nil {
				t.Errorf("Failed to write event: %v", err)
				return
			}
			t.Log("Wrote event on the event socket")
			time.Sleep(10 * time.Millisecond)
		}
	}()
	return serverDone
}
