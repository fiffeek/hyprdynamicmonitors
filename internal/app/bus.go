package app

import (
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/sirupsen/logrus"
)

func getBus(connectToSessionBus bool) (*dbus.Conn, error) {
	var conn *dbus.Conn
	var err error
	if connectToSessionBus {
		logrus.Debug("Trying to connect to session bus")
		conn, err = dbus.ConnectSessionBus()
	} else {

		logrus.Debug("Trying to connect to system bus")
		conn, err = dbus.ConnectSystemBus()
	}

	if err != nil {
		return nil, fmt.Errorf("cant init dbus conn: %w", err)
	}

	return conn, nil
}
