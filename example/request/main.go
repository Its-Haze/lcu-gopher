package main

import (
	"fmt"
	"io"
	"log"

	"github.com/its-haze/lcu-gopher"
)

func main() {
	// Create client and connect to LCU
	client, err := lcu.NewClient(lcu.DefaultConfig())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	fmt.Println("Successfully connected to LCU Api")

	// Get current summoner information
	// resp, err := client.Get("/lol-summoner/v1/current-summoner")
	resp, err := client.Get("/lol-chat/v1/friends")
	if err != nil {
		log.Fatalf("Failed to get summoner info: %v", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	// Print the raw response
	fmt.Printf("\nCurrent Summoner Info:\n%s\n", string(body))
}
