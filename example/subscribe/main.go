package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/its-haze/lcu-gopher"
)

// handleSummonerUpdate processes summoner profile update events
func handleSummonerUpdate(event *lcu.Event) {
	if data, ok := event.Data.(map[string]interface{}); ok {
		if gameName, ok := data["gameName"].(string); ok {
			fmt.Printf("%s updated their summoner profile\n", gameName)
		}
	}
}

func main() {
	config := lcu.DefaultConfig()
	config.AwaitConnection = true // Wait for LCU to start

	client, err := lcu.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect()
	if err := client.Connect(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully connected to LCU Api")

	if err := client.Subscribe("/lol-summoner/v1/current-summoner", handleSummonerUpdate, lcu.EventTypeUpdate); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Change your profile picture... (Press Ctrl+C to exit)")

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
