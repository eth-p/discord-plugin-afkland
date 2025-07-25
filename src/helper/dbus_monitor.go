package main

import (
	"context"
	"fmt"

	"github.com/godbus/dbus/v5"
)

// DBusMonitor monitors DBus for messages.
// This has its own connection, as using BecomeMonitor prevents sending messages.
//
// https://dbus.freedesktop.org/doc/dbus-specification.html#bus-messages-become-monitor
type DBusMonitor struct {
	conn  *dbus.Conn
	rules []string
	spmc  *SPMC[*dbus.Message]
	ch    chan *dbus.Message
}

func NewDBusMonitor() (*DBusMonitor, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, fmt.Errorf("connecting to DBus session bus: %w", err)
	}

	eavesdropCh := make(chan *dbus.Message, 32)

	return &DBusMonitor{
		conn:  conn,
		ch:    eavesdropCh,
		spmc:  NewSPMC(eavesdropCh),
		rules: []string{},
	}, nil
}

func (m *DBusMonitor) AddMatchRules(rules []string) error {
	m.rules = append(m.rules, rules...)

	m.conn.Eavesdrop(nil)
	defer m.conn.Eavesdrop(m.ch)

	flag := uint(0)
	call := m.conn.BusObject().Call("org.freedesktop.DBus.Monitoring.BecomeMonitor", 0, m.rules, flag)
	if call.Err != nil {
		return fmt.Errorf("BecomeMonitor: %w", call.Err)
	}

	return nil
}

func (m *DBusMonitor) Listen() (<-chan *dbus.Message, context.CancelFunc) {
	return m.spmc.Consumer(cap(m.ch))
}

func (m *DBusMonitor) Close() error {
	close(m.ch)
	m.spmc.Close()
	return m.conn.Close()
}
