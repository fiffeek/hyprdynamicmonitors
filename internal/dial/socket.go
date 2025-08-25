// Package dial provides unix socket helpers.
package dial

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/sirupsen/logrus"
)

func GetUnixSocketConnection(ctx context.Context, socketPath string) (net.Conn, func(), error) {
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("hyprland command socket not found at %s", socketPath)
	}

	d := &net.Dialer{}
	conn, err := d.DialContext(ctx, "unix", socketPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to socket: %w", err)
	}

	return conn, func() {
		if err := conn.Close(); err != nil {
			logrus.WithError(err).Debug("Failed to close connection")
		}
	}, nil
}

type SocketJSONResponse interface {
	Validate() error
}

func SyncQuerySocket[T SocketJSONResponse](conn net.Conn, command string) (T, error) {
	var zero T

	_, err := conn.Write([]byte(command))
	if err != nil {
		return zero, fmt.Errorf("failed to command: %w", err)
	}

	response, err := io.ReadAll(conn)
	if err != nil {
		return zero, fmt.Errorf("failed to response: %w", err)
	}

	logrus.WithFields(logrus.Fields{"response": string(response)}).Debug("ipc response")

	var res T
	if err := utils.UnmarshalResponse(response, &res); err != nil {
		return zero, fmt.Errorf("failed to parse response: %w", err)
	}

	if err := res.Validate(); err != nil {
		return zero, fmt.Errorf("failed to validate response: %w", err)
	}

	return res, nil
}
