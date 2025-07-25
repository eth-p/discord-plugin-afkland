package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

const (
	listenAddress = "127.0.0.1:16738"
	listenPath    = "/afkland-native-helper"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return r.Header.Get("Origin") == "https://discord.com"
	},
}

var (
	monitor     *DBusMonitor
	notifyAfk   chan bool   = make(chan bool, 32)
	notifiedAfk *SPMC[bool] = NewSPMC(notifyAfk)
)

func main() {
	var err error
	defer close(notifyAfk)
	defer notifiedAfk.Close()

	log.Println("Attaching to DBus...")
	monitor, err = NewDBusMonitor()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	go ListenForScreenSaver(monitor, notifyAfk)

	// Start the websocket server which the BetterDiscord plugin
	// will connect to.
	log.Printf("Starting websocket server at %s...\n", listenAddress)
	http.HandleFunc(listenPath, helper)
	err = http.ListenAndServe(listenAddress, nil)
	if err != nil {
		fmt.Println("Failed to listen for websocket connections.")
		fmt.Println(err)
		os.Exit(1)
	}
}

func helper(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade websocket connection: %v\n", err)
		return
	}

	log.Println("Websocket connection established.")
	defer c.Close()

	// Discard all incoming messages.
	go func() {
		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				log.Printf("Websocket read error: %v", err)
				break
			}
		}
	}()

	// Send AFK events.
	afkEventCh, close := notifiedAfk.Consumer(16)
	defer close()
	for {
		afkStatus, ok := <-afkEventCh
		if !ok {
			return
		}

		afkStatusMessage, err := json.Marshal(afkStatus)
		if err != nil {
			log.Printf("Failed to encode AFK status message: %v", err)
			return
		}

		err = c.WriteMessage(websocket.TextMessage, afkStatusMessage)
		if err != nil {
			log.Printf("Failed to send AFK status message: %v", err)
			return
		}
	}
}
