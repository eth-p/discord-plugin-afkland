package main

import (
	"fmt"
	"log"

	"github.com/godbus/dbus/v5"
)

func ListenForScreenSaver(monitor *DBusMonitor, afkStatus chan<- bool) {
	err := monitor.AddMatchRules([]string{
		"type='signal',interface='org.freedesktop.ScreenSaver'",
	})

	if err != nil {
		panic(fmt.Errorf("failed to listen for ScreenSaver DBus signal: %w", err))
	}

	events, close := monitor.Listen()
	defer close()

	for {
		evt := <-events

		if evt.Headers[dbus.FieldInterface].String() != `"org.freedesktop.ScreenSaver"` {
			continue
		}

		if evt.Headers[dbus.FieldMember].String() != `"ActiveChanged"` {
			continue
		}

		status := evt.Body[0].(bool)
		log.Printf("ScreenSaver status updated to: %v", status)
		afkStatus <- status
	}
}
