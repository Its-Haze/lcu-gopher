# lcu-gopher

A Go library for interacting with the League of Legends Client API (LCU). This library provides a simple and efficient way to connect to the League Client, make HTTP requests, and subscribe to WebSocket events.

## Features

- 🔌 Automatic LCU process detection and connection
- 🔄 WebSocket event subscription with type filtering
- 🌐 HTTP request methods (GET, POST, PUT, DELETE)
- 🔍 Configurable logging with debug mode
- ⏱️ Customizable timeouts and polling intervals
- 🔒 Automatic authentication handling
- 🗂️ Flexible League Client installation path detection

## Installation

```bash
go get github.com/its-haze/lcu-gopher
```

## Quick Start

```go
package main

import (
	"fmt"
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
	summoner, err := client.GetCurrentSummoner()
	if err != nil {
		log.Fatalf("Failed to get summoner info: %v", err)
	}

	fmt.Printf("Your Level is: %d\n", summoner.SummonerLevel)
}
```

## LCU API Documentation

The League Client API provides a comprehensive set of endpoints for interacting with the game client. You can find the complete API documentation, including all available endpoints, request/response schemas, and WebSocket events at:

[LCU API Documentation](https://www.mingweisamuel.com/lcu-schema/tool/#/)

When using this library, refer to the documentation to understand the full capabilities of the LCU API and how to structure your requests and handle responses.

## Configuration

The library provides several configuration options through the `Config` struct:

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

### League Client Path Detection

The library automatically detects the League Client installation path using multiple methods:

1. **Custom Path**: If `LeaguePath` is set in the config, it will be used first
2. **Lockfile Detection**: Searches for the lockfile in common installation locations
3. **Process Detection**: Finds the running League Client process and extracts its path

Supported installation paths:

#### Windows
- `C:\Riot Games\League of Legends`
- `C:\Program Files\Riot Games\League of Legends`
- `C:\Program Files (x86)\Riot Games\League of Legends`
- Custom paths on other drives (D:, E:, F:, G:)

#### macOS
- `/Applications/League of Legends.app/Contents/LoL`
- `$HOME/Applications/League of Legends.app/Contents/LoL`

#### Linux
- WSL2: Windows paths mounted under `/mnt/[drive]/`

## Examples

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

### Custom Logging

Implement the `Logger` interface for custom logging:

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

## Examples Directory

The repository includes several example applications:

- `example/request/main.go`: Demonstrates making HTTP requests to the LCU
- `example/subscribe/main.go`: Shows how to subscribe to LCU events

## Error Handling

The library returns descriptive errors for common issues:
- Connection failures
- Invalid event types
- Authentication errors
- Timeout errors

## Common Issues

### Duplicate WebSocket Events

The League Client API may send multiple events for the same state change. For example, when starting or stopping queue, you might receive multiple "Matchmaking" or "Lobby" events:

```
Game phase changed to: Lobby
Game phase changed to: Matchmaking
Game phase changed to: Matchmaking
Game phase changed to: Matchmaking
Game phase changed to: Lobby
Game phase changed to: Lobby
```

This is normal behavior from the League Client API and not a bug in this library. If you need to handle this in your application, you can implement your own event deduplication logic.

Example solution:
```go
type GamePhaseHandler struct {
    lastPhase GamePhase
    lastTime  time.Time
    mu        sync.Mutex
}

func (h *GamePhaseHandler) Handle(phase GamePhase) {
    h.mu.Lock()
    defer h.mu.Unlock()

    // Ignore duplicate events within 500ms
    if phase == h.lastPhase && time.Since(h.lastTime) < 500*time.Millisecond {
        return
    }

    h.lastPhase = phase
    h.lastTime = time.Now()
    fmt.Printf("Game phase changed to: %s\n", phase)
}
```

### Connection Timeouts

If you're experiencing connection timeouts, try:
1. Increasing the `Timeout` value in the config
2. Ensuring the League Client is running and fully loaded
3. Checking if your firewall is blocking the connection

### WebSocket Disconnections

The WebSocket connection might disconnect unexpectedly. The library handles reconnection automatically, but you might want to implement your own reconnection logic for more control:

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

The League Client API has rate limits. If you're making many requests, you might want to implement rate limiting in your application:

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

### Process Detection Issues

If the library fails to detect the League Client process:
1. Ensure the League Client is running
2. Check if the League Client is installed in a non-standard location
3. Use the `LeaguePath` config option to specify the installation path
4. Enable debug logging to see which paths are being checked

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 