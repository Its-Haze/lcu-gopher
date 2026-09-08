# lcu-gopher 🎮

A Go library for interacting with the League of Legends Client API (LCU). It finds the running client's credentials, makes authenticated HTTP requests to it, and dispatches its WebSocket events to your handlers.

[![Go Report Card](https://goreportcard.com/badge/github.com/its-haze/lcu-gopher)](https://goreportcard.com/report/github.com/its-haze/lcu-gopher)
[![GoDoc](https://godoc.org/github.com/its-haze/lcu-gopher?status.svg)](https://godoc.org/github.com/its-haze/lcu-gopher)

## Features

- **Credential discovery**: Reads the `LeagueClientUx` process command line, falling back to the client's lockfile, and health-checks either before accepting it
- **WebSocket events**: Subscribe to game events by endpoint and event type
- **HTTP methods**: GET, POST, PUT, and DELETE against any LCU endpoint
- **Optional wait-for-client**: Poll until the client is up and answering, instead of failing immediately
- **Configurable logging**: Console logging, or one log file per endpoint
- **Typed helpers**: Structs and getters for the common endpoints (summoner, gameflow, lobby, champ select, ranked, friends)

Dependencies are `github.com/gorilla/websocket` and `github.com/shirou/gopsutil/v4`. Windows is the supported target: process discovery matches `LeagueClientUx.exe` by name. Lockfile paths for macOS and WSL2 are still in the code but are not tested, and Vanguard blocks League on Linux anyway.

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
	// Resolves credentials immediately, so this fails if the client isn't running
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

## Configuration

```go
config := &lcu.Config{
	PollInterval:    2 * time.Second,    // How often to check for the LCU process
	Timeout:         30 * time.Second,   // HTTP request timeout
	Logger:          &MyLogger{},        // Custom logger (must not be nil)
	AwaitConnection: false,              // Whether to wait for LCU to start
	Debug:           false,              // Enable debug logging
	LogDir:          "",                 // Directory for endpoint-specific logs
	LeaguePath:      "",                 // Custom path to League installation
	ProcessLister:   nil,                // Overrides process discovery; nil reads the real process table
}
```

Start from `DefaultConfig()` and change what you need:
```go
config := lcu.DefaultConfig()
config.Debug = true
```

Building a `Config` by hand is fine, but `Logger` cannot be nil. Credential discovery and the wait loop call it directly with no nil check, so a bare `&lcu.Config{}` panics. `DefaultConfig()` fills in a console logger for you.

Setting `Debug` without `LogDir` defaults `LogDir` to `logs`, created relative to the process's working directory. That is rarely where you want it, and it fails outright if the working directory is not writable, which takes `NewClient` down with it. Set `LogDir` explicitly.

## Examples

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
err := client.Subscribe("/lol-summoner/v1/current-summoner", handleSummonerUpdate, lcu.EventTypeUpdate)

// Subscribe to all events
err := client.SubscribeToAll(handleAllEvents)
```

At least one event type is required. `Subscribe` returns an error if you pass none, or if you pass anything other than `Create`, `Update`, or `Delete`.

Each handler runs in its own goroutine per event, so handlers must be safe to call concurrently and must not assume events arrive in order.

Check out the [examples directory](example/) for more detailed examples:
- [Basic HTTP Requests](example/request/main.go)
- [Event Subscription](example/subscribe/main.go)
- [Game Flow Phase Monitoring](example/gameflowphase/main.go)

## LCU API Documentation

The League Client API provides a comprehensive set of endpoints. You can find the complete API documentation at:

[Swagger LCU API Documentation](https://www.mingweisamuel.com/lcu-schema/tool/#/)

## Common Use Cases

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

### Waiting for the Client to Start
`NewClient` resolves credentials before it returns, so it fails immediately when the client isn't running. Set `AwaitConnection` to poll instead. The blocking happens in `NewClient`, not in `Connect`:

```go
config := lcu.DefaultConfig()
config.AwaitConnection = true
config.PollInterval = 2 * time.Second

// Blocks until the client is up and answering /lol-summoner/v1/current-summoner
client, err := lcu.NewClient(config)
```

## ⚠️ Limitations

Read this before building anything long-running on top of the library.

**A `Client` is single-use.** `Disconnect()` closes an internal channel, so calling it twice panics, and a client that has been disconnected cannot be reconnected. Build a new one with `NewClient`.

**There is no automatic reconnection.** If the WebSocket read fails, the listener goroutine logs the error and stops. Nothing retries, and no callback fires.

**There is no disconnect detection.** Nothing notices the League Client closing mid-session. Requests start failing and events stop arriving, but the library will not tell you it happened.

To survive the client restarting, supervise it from the outside: watch for the `LeagueClientUx` process yourself and build a fresh client when it comes back. Credentials are only read during `NewClient`, and the port and auth token change on every client restart, so an old client would be talking to a dead port anyway.

```go
func connect() (*lcu.Client, error) {
	config := lcu.DefaultConfig()
	config.AwaitConnection = true // blocks here until the client is up

	client, err := lcu.NewClient(config)
	if err != nil {
		return nil, err
	}
	if err := client.Connect(); err != nil {
		return nil, err
	}
	// Subscriptions live on the WebSocket connection, so re-register them here
	return client, nil
}
```

**No rate limiting.** The library sends whatever you ask it to send. The LCU tolerates a lot for a local API, but throttle on your side if you are polling an endpoint in a tight loop.

**TLS verification is disabled.** The League Client serves a self-signed certificate, so both the HTTP client and the WebSocket dialer use `InsecureSkipVerify`. That is unavoidable for the LCU and harmless over loopback, but worth knowing it is there.

## Troubleshooting

### Connection failures
1. Make sure the League Client is running and fully loaded, not just launching. `LeagueClientUx.exe` exists for a while before the API answers on it
2. Turn on `Debug` to see which discovery method was tried, which one produced credentials, and what the health check returned
3. A non-default install directory needs no configuration: the process scan reads the client's own `--install-directory`. `LeaguePath` is only a manual override

### Connection timeouts
1. Raise `Timeout` in the config
2. Check whether a firewall is blocking the loopback connection

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
