# lcu-gopher 🎮

A powerful Go library for interacting with the League of Legends Client API (LCU). This library provides a simple and efficient way to connect to the League Client, make HTTP requests, and subscribe to WebSocket events.

[![Go Report Card](https://goreportcard.com/badge/github.com/its-haze/lcu-gopher)](https://goreportcard.com/report/github.com/its-haze/lcu-gopher)
[![GoDoc](https://godoc.org/github.com/its-haze/lcu-gopher?status.svg)](https://godoc.org/github.com/its-haze/lcu-gopher)

## 🌟 Features

- 🔌 **Automatic Connection**: Automatically detects and connects to the League Client
- 🔄 **WebSocket Support**: Subscribe to real-time game events and updates
- 🌐 **HTTP Methods**: Full support for GET, POST, PUT, and DELETE requests
- 🔍 **Debug Mode**: Configurable logging with detailed debug information
- ⏱️ **Customizable**: Adjustable timeouts and polling intervals
- 🔒 **Secure**: Built-in authentication handling
- 🗂️ **Flexible**: Supports multiple League Client installation paths
- 📝 **Well Documented**: Comprehensive API documentation and examples

## 📦 Installation

```bash
go get github.com/its-haze/lcu-gopher
```

## 🚀 Quick Start

Here's a simple example to get you started:

```go
package main

import (
	"fmt"
	"log"

	"github.com/its-haze/lcu-gopher"
)

func main() {
	// Create a new client with default configuration
	client, err := lcu.NewClient(lcu.DefaultConfig())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Connect to the League Client
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	// Get current summoner information
	summoner, err := client.GetCurrentSummoner()
	if err != nil {
		log.Fatalf("Failed to get summoner info: %v", err)
	}

	fmt.Printf("Welcome, %s! (Level %d)\n", summoner.GameName, summoner.SummonerLevel)
}
```

## ⚙️ Configuration

The library is highly configurable through the `Config` struct:

```go
config := &lcu.Config{
	PollInterval:    2 * time.Second,    // How often to check for LCU process
	Timeout:         30 * time.Second,   // HTTP request timeout
	Logger:          nil,                // Custom logger (optional)
	AwaitConnection: false,              // Whether to wait for LCU to start
	Debug:           false,              // Enable debug logging
	LogDir:          "",                 // Directory for endpoint-specific logs
	LeaguePath:      "",                 // Custom path to League installation
}
```

Use `DefaultConfig()` for default settings:
```go
config := lcu.DefaultConfig()
config.Debug = true  // Enable debug logging
```

## 📚 Examples

The repository includes several example applications to help you get started:

### Making HTTP Requests
```go
// GET request
resp, err := client.Get("/lol-summoner/v1/current-summoner")

// POST request with body
body := strings.NewReader(`{"key": "value"}`)
resp, err := client.Post("/some-endpoint", body)

// PUT request
resp, err := client.Put("/some-endpoint", body)

// DELETE request
resp, err := client.Delete("/some-endpoint")
```

### Subscribing to Events
```go
// Handler function
func handleSummonerUpdate(event *lcu.Event) {
	if data, ok := event.Data.(map[string]interface{}); ok {
		if gameName, ok := data["gameName"].(string); ok {
			fmt.Printf("%s updated their summoner profile\n", gameName)
		}
	}
}

// Subscribe to specific event types
err := client.Subscribe("/lol-summoner/v1/current-summoner", handleSummonerUpdate, "Update")

// Subscribe to all events
err := client.SubscribeToAll(handleAllEvents)
```

Check out the [examples directory](example/) for more detailed examples:
- [Basic HTTP Requests](example/request/main.go)
- [Event Subscription](example/subscribe/main.go)
- [Game Flow Phase Monitoring](example/gameflowphase/main.go)

## 🔍 LCU API Documentation

The League Client API provides a comprehensive set of endpoints. You can find the complete API documentation at:

[Swagger LCU API Documentation](https://www.mingweisamuel.com/lcu-schema/tool/#/)


## 🛠️ Common Use Cases

### Custom Logging
```go
type MyLogger struct{}

func (l *MyLogger) Info(endpoint, msg string, args ...interface{}) {
	// Your logging implementation
}

func (l *MyLogger) Error(endpoint, msg string, args ...interface{}) {
	// Your logging implementation
}

func (l *MyLogger) Debug(endpoint, msg string, args ...interface{}) {
	// Your logging implementation
}

// Use custom logger
config := lcu.DefaultConfig()
config.Logger = &MyLogger{}
```

### Handling Game Phases
```go
client.SubscribeToGamePhase(func(phase lcu.GamePhase) {
	switch phase {
	case lcu.GamePhaseLobby:
		fmt.Println("In lobby")
	case lcu.GamePhaseMatchmaking:
		fmt.Println("In queue")
	case lcu.GamePhaseChampSelect:
		fmt.Println("In champion select")
	case lcu.GamePhaseInProgress:
		fmt.Println("Game in progress")
	}
})
```

## ⚠️ Common Issues

### Connection Timeouts
If you're experiencing connection timeouts:
1. Increase the `Timeout` value in the config
2. Ensure the League Client is running and fully loaded
3. Check if your firewall is blocking the connection

### WebSocket Disconnections
The library handles reconnection automatically, but you can implement custom reconnection logic:
```go
func handleDisconnection(client *lcu.Client) {
	for {
		if err := client.Connect(); err != nil {
			log.Printf("Failed to reconnect: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
}
```

### Rate Limiting
The League Client API has rate limits. Implement rate limiting in your application if needed:
```go
type RateLimiter struct {
	tokens     int
	maxTokens  int
	lastRefill time.Time
	mu         sync.Mutex
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)
	refillAmount := int(elapsed / time.Second)
	
	if refillAmount > 0 {
		rl.tokens = min(rl.maxTokens, rl.tokens+refillAmount)
		rl.lastRefill = now
	}

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	return false
}
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 