package main

import (
	"fmt"
	"log"

	"github.com/its-haze/lcu-gopher"
)

func main() {
	// Create a new client with default configuration
	client, err := lcu.NewClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Connect to the LCU
	if err := client.Connect(); err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect()

	// Get current summoner information
	summoner, err := client.GetCurrentSummoner()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Logged in as: %s (Level %d)\n", summoner.DisplayName, summoner.SummonerLevel)

	// Subscribe to game phase changes
	client.SubscribeToGamePhase(func(phase lcu.GamePhase) {
		fmt.Printf("Game phase changed to: %s\n", phase)
	})

	// Keep the program running
	select {}
}
