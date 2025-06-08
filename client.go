package lcu

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client represents a connection to the League Client API.
// It manages HTTP and WebSocket connections, handles authentication,
// and provides methods for subscribing to events and making requests.
//
// The client uses a WAMP protocol for event communication, which is a
// publish-subscribe pattern that allows for efficient event handling.
//
// The client supports both HTTP and WebSocket connections, and can be
// configured with various options for logging, debugging, and event handling.
//
// The client is designed to be used in a multi-threaded environment,
// and provides methods for subscribing to events and making requests.
type Client struct {
	credentials *Credentials
	httpClient  *http.Client
	wsConn      *websocket.Conn
	wsLock      sync.RWMutex
	eventMux    sync.RWMutex
	handlers    map[string][]EventHandler
	done        chan struct{}
	logger      Logger
	config      *Config
}

// Credentials represents the authentication credentials for the League Client API.
type Credentials struct {
	Port     int    `json:"port"`
	Password string `json:"password"`
	Protocol string `json:"protocol"`
}

// EventHandler represents a function that handles LCU events
type EventHandler func(event *Event)

// Event represents a LCU websocket event.
// It contains information about the event, including the event type, URI, and data.
//
// The event type is the type of event that was received, such as "Create", "Update", or "Delete".
// The URI is the URI of the endpoint that the event was received from.
// The data is the data of the event, which is typically a JSON object.
type Event struct {
	EventType string      `json:"eventType"`
	URI       string      `json:"uri"`
	Data      interface{} `json:"data"`
}

// Logger interface for logging (users can implement their own)
type Logger interface {
	Info(endpoint, msg string, args ...interface{})
	Error(endpoint, msg string, args ...interface{})
	Debug(endpoint, msg string, args ...interface{})
}

// Config represents the configuration for the LCU client.
type Config struct {
	PollInterval    time.Duration // How often to check for LCU process
	Timeout         time.Duration // HTTP request timeout
	Logger          Logger        // Custom logger
	AwaitConnection bool          // Whether to wait for LCU to start
	Debug           bool          // Whether to enable debug logging
	LogDir          string        // Directory to store endpoint-specific log files

	// Custom path to League of Legends installation
	// Example: "C:\\Riot Games\\League of Legends"
	LeaguePath string
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		PollInterval:    2 * time.Second,
		Timeout:         30 * time.Second,
		Logger:          &defaultLogger{},
		AwaitConnection: false,
		Debug:           false,
		LogDir:          "", // Empty by default, will be set if debug is enabled
		LeaguePath:      "", // Empty by default, will be auto-detected
	}
}

// Default logger implementation
type defaultLogger struct {
	debug bool
}

func (l *defaultLogger) log(level, endpoint, msg string, args ...interface{}) {
	// Skip debug logs if debug mode is not enabled
	if level == "DEBUG" && !l.debug {
		return
	}

	// Format the message with arguments
	formattedMsg := fmt.Sprintf(msg, args...)

	// Add timestamp and level
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	var logMsg string
	if level == "DEBUG" {
		// Create a separator line based on the endpoint length
		separator := strings.Repeat("-", len(endpoint)+4)

		// Format the log message with better grouping for DEBUG
		logMsg = fmt.Sprintf("\n[%s] [%s]\n%s\nEndpoint: %s\nMessage: %s\n%s\n",
			timestamp,
			level,
			separator,
			endpoint,
			formattedMsg,
			separator)
	} else {
		// Simple format for INFO and ERROR
		logMsg = fmt.Sprintf("[%s] [%s] %s\n",
			timestamp,
			level,
			formattedMsg)
	}

	// Log to console
	fmt.Print(logMsg)
}

func (l *defaultLogger) Info(endpoint, msg string, args ...interface{}) {
	l.log("INFO", endpoint, msg, args...)
}

func (l *defaultLogger) Error(endpoint, msg string, args ...interface{}) {
	l.log("ERROR", endpoint, msg, args...)
}

func (l *defaultLogger) Debug(endpoint, msg string, args ...interface{}) {
	l.log("DEBUG", endpoint, msg, args...)
}

// EndpointLogger handles logging to endpoint-specific files
type EndpointLogger struct {
	logDir    string
	logFiles  map[string]*os.File
	fileMutex sync.RWMutex
}

// NewEndpointLogger creates a new endpoint-specific logger
func NewEndpointLogger(logDir string) (*EndpointLogger, error) {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	return &EndpointLogger{
		logDir:   logDir,
		logFiles: make(map[string]*os.File),
	}, nil
}

func (l *EndpointLogger) getLogFile(endpoint string) (*os.File, error) {
	l.fileMutex.RLock()
	if file, exists := l.logFiles[endpoint]; exists {
		l.fileMutex.RUnlock()
		return file, nil
	}
	l.fileMutex.RUnlock()

	// Create new file if it doesn't exist
	l.fileMutex.Lock()
	defer l.fileMutex.Unlock()

	// Double check after acquiring write lock
	if file, exists := l.logFiles[endpoint]; exists {
		return file, nil
	}

	// Create sanitized filename from endpoint
	filename := strings.ReplaceAll(endpoint, "/", "_")
	if filename == "" {
		filename = "root"
	}
	filepath := filepath.Join(l.logDir, filename+".log")

	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file for endpoint %s: %w", endpoint, err)
	}

	l.logFiles[endpoint] = file
	return file, nil
}

func (l *EndpointLogger) log(level, endpoint, msg string, args ...interface{}) {
	// Format the message with arguments
	formattedMsg := fmt.Sprintf(msg, args...)

	// Add timestamp and level
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	// Create a separator line based on the endpoint length
	separator := strings.Repeat("-", len(endpoint)+4)

	// Format the log message with better grouping
	logMsg := fmt.Sprintf("\n[%s] [%s]\n%s\nEndpoint: %s\nMessage: %s\n%s\n",
		timestamp,
		level,
		separator,
		endpoint,
		formattedMsg,
		separator)

	// Log to console
	fmt.Print(logMsg)

	// Log to endpoint-specific file
	if file, err := l.getLogFile(endpoint); err == nil {
		file.WriteString(logMsg)
	}
}

func (l *EndpointLogger) Info(endpoint, msg string, args ...interface{}) {
	l.log("INFO", endpoint, msg, args...)
}

func (l *EndpointLogger) Error(endpoint, msg string, args ...interface{}) {
	l.log("ERROR", endpoint, msg, args...)
}

func (l *EndpointLogger) Debug(endpoint, msg string, args ...interface{}) {
	l.log("DEBUG", endpoint, msg, args...)
}

func (l *EndpointLogger) Close() {
	l.fileMutex.Lock()
	defer l.fileMutex.Unlock()

	for _, file := range l.logFiles {
		file.Close()
	}
	l.logFiles = make(map[string]*os.File)
}

// NewClient creates a new LCU client with the specified configuration.
// If no configuration is provided, it uses the default configuration.
//
// The function performs the following:
//   - Sets up file logging if debug mode is enabled
//   - Creates an endpoint-specific logger if LogDir is specified
//   - Finds LCU credentials using the provided configuration
//   - Initializes an HTTP client with TLS configuration for self-signed certificates
//   - Sets up event handlers and logging infrastructure
//
// Parameters:
//   - config: Configuration options for the client (optional)
//
// Returns:
//   - *Client: A new LCU client instance
//   - error: Any error that occurred during client creation
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Set up file logging only if debug mode is enabled
	if config.Debug && config.LogDir == "" {
		config.LogDir = "logs" // Set default log directory only if debug is enabled
	}

	// Set up file logging if configured
	if config.LogDir != "" {
		logger, err := NewEndpointLogger(config.LogDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create endpoint logger: %w", err)
		}

		// If using default logger, replace it with the endpoint logger
		if _, ok := config.Logger.(*defaultLogger); ok {
			config.Logger = logger
		}
	} else if defaultLogger, ok := config.Logger.(*defaultLogger); ok {
		// Set debug mode on the default logger
		defaultLogger.debug = config.Debug
	}

	credentials, err := findCredentials(config)
	if err != nil {
		return nil, fmt.Errorf("failed to find LCU credentials: %w", err)
	}

	client := &Client{
		credentials: credentials,
		httpClient: &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // LCU uses self-signed cert
				},
			},
		},
		handlers: make(map[string][]EventHandler),
		done:     make(chan struct{}),
		logger:   config.Logger,
		config:   config,
	}

	return client, nil
}

// Connect establishes a connection to the League Client API by:
// 1. Testing the HTTP connection to verify basic connectivity
// 2. Establishing a WebSocket connection for real-time event handling
// Returns an error if either connection attempt fails
func (c *Client) Connect() error {
	// Test HTTP connection first
	if err := c.testConnection(); err != nil {
		return fmt.Errorf("failed to establish HTTP connection: %w", err)
	}

	// Establish WebSocket connection for events
	if err := c.connectWebSocket(); err != nil {
		return fmt.Errorf("failed to establish WebSocket connection: %w", err)
	}

	c.logger.Debug("connection", "Successfully connected to LCU on port %d", c.credentials.Port)
	return nil
}

// Disconnect closes all connections by:
// 1. Closing the WebSocket connection
// 2. Closing any associated log files
// Returns an error if the WebSocket connection fails to close
func (c *Client) Disconnect() error {
	close(c.done)

	c.wsLock.Lock()
	if c.wsConn != nil {
		c.wsConn.Close()
		c.wsConn = nil
	}
	c.wsLock.Unlock()

	// Close log files if using endpoint logger
	if endpointLogger, ok := c.logger.(interface {
		Close()
	}); ok {
		endpointLogger.Close()
	}

	return nil
}

// Request sends an HTTP request to the specified endpoint with the given method and body.
// It handles authentication, logging, and debug mode.
//
// Parameters:
//   - method: The HTTP method to use (e.g., "GET", "POST")
//   - endpoint: The API endpoint to request (e.g., "/lol-summoner/v1/current-summoner")
//   - body: The request body (optional)
//
// Returns:
//   - *http.Response: The HTTP response from the request
//   - error: Any error that occurred during the request
func (c *Client) Request(method, endpoint string, body io.Reader) (*http.Response, error) {
	baseURL := fmt.Sprintf("https://127.0.0.1:%d", c.credentials.Port)
	reqURL, err := url.JoinPath(baseURL, endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}

	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, err
	}

	// Add authentication header
	auth := base64.StdEncoding.EncodeToString([]byte("riot:" + c.credentials.Password))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")

	// Debug logging for request
	if c.config.Debug {
		c.logger.Debug(endpoint, "Making %s request to %s", method, reqURL)
		if body != nil {
			bodyBytes, _ := io.ReadAll(body)
			c.logger.Debug(endpoint, "Request body: %s", string(bodyBytes))
			// Reset body reader for actual request
			body = bytes.NewReader(bodyBytes)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Debug logging for response
	if c.config.Debug {
		bodyBytes, _ := io.ReadAll(resp.Body)
		c.logger.Debug(endpoint, "Response status: %s", resp.Status)
		c.logger.Debug(endpoint, "Response body: %s", string(bodyBytes))
		// Reset body reader for actual response
		resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	return resp, nil
}

// Get performs a GET request
func (c *Client) Get(endpoint string) (*http.Response, error) {
	return c.Request("GET", endpoint, nil)
}

// Post performs a POST request
func (c *Client) Post(endpoint string, body io.Reader) (*http.Response, error) {
	return c.Request("POST", endpoint, body)
}

// Put performs a PUT request
func (c *Client) Put(endpoint string, body io.Reader) (*http.Response, error) {
	return c.Request("PUT", endpoint, body)
}

// Delete performs a DELETE request
func (c *Client) Delete(endpoint string) (*http.Response, error) {
	return c.Request("DELETE", endpoint, nil)
}

// Valid event types for LCU
var validEventTypes = map[string]bool{
	"Create": true,
	"Update": true,
	"Delete": true,
}

// Subscribe registers an event handler for a specific endpoint and event types.
// The handler will be called when events of the specified types (Create, Update, Delete)
// are received for the given endpoint. The handler is also registered for the general
// event bus (OnJsonApiEvent) to ensure no events are missed.
//
// Parameters:
//   - endpoint: The API endpoint to subscribe to (e.g., "/lol-gameflow/v1/session")
//   - handler: The function that will be called when matching events are received
//   - eventTypes: One or more event types to filter for (Create, Update, Delete)
//
// Returns an error if:
//   - No event types are specified
//   - An invalid event type is provided
//   - Failed to send subscription message via WebSocket
func (c *Client) Subscribe(endpoint string, handler EventHandler, eventTypes ...EventType) error {
	// Validate event types if provided
	if len(eventTypes) > 0 {
		for _, eventType := range eventTypes {
			if !validEventTypes[string(eventType)] {
				return fmt.Errorf("invalid event type: %s. Valid types are: Create, Update, Delete", eventType)
			}
		}
	} else {
		return fmt.Errorf("at least one event type must be specified. Valid types are: Create, Update, Delete")
	}

	c.eventMux.Lock()
	defer c.eventMux.Unlock()

	// Create a wrapper handler that checks event types
	wrappedHandler := func(event *Event) {
		// Check if event type matches any of the specified types
		for _, eventType := range eventTypes {
			if event.EventType == string(eventType) {
				handler(event)
				return
			}
		}
	}

	// Add handler to both the specific endpoint and the event bus
	c.handlers[endpoint] = append(c.handlers[endpoint], wrappedHandler)
	c.handlers["OnJsonApiEvent"] = append(c.handlers["OnJsonApiEvent"], wrappedHandler)

	// Subscribe to both the specific endpoint and the general event bus
	subscriptions := []string{endpoint, "OnJsonApiEvent"}
	for _, uri := range subscriptions {
		message := []interface{}{5, uri}
		if err := c.sendWebSocketMessage(message); err != nil {
			return fmt.Errorf("failed to send subscription message for %s: %w", uri, err)
		}
	}

	return nil
}

// Unsubscribe removes an event handler for a specific endpoint.
// It sends a WAMP unsubscription message to the WebSocket connection.
//
// Parameters:
//   - endpoint: The API endpoint to unsubscribe from (e.g., "/lol-gameflow/v1/session")
//
// Returns an error if:
//   - Failed to send unsubscription message via WebSocket
//   - Failed to delete the handler from the internal map
func (c *Client) Unsubscribe(endpoint string) error {
	c.eventMux.Lock()
	delete(c.handlers, endpoint)
	c.eventMux.Unlock()

	// Send unsubscription message via WebSocket (WAMP protocol)
	return c.sendWebSocketMessage([]interface{}{6, endpoint})
}

// SubscribeToAll registers an event handler for all events received from the event bus.
// This is useful for handling events that don't have a specific endpoint.
//
// Parameters:
//   - handler: The function that will be called when any event is received
//
// Returns an error if:
//   - Failed to send subscription message via WebSocket
//   - Failed to add the handler to the internal map
func (c *Client) SubscribeToAll(handler EventHandler) error {
	return c.Subscribe("/", handler, EventTypeCreate, EventTypeUpdate, EventTypeDelete)
}

// Private methods

func (c *Client) testConnection() error {
	resp, err := c.Get("/lol-summoner/v1/current-summoner")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}

func (c *Client) connectWebSocket() error {
	wsURL := fmt.Sprintf("wss://127.0.0.1:%d/", c.credentials.Port)

	if c.config.Debug {
		c.logger.Debug("websocket", "Connecting to WebSocket at %s", wsURL)
	}

	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		Subprotocols: []string{"wamp"},
	}

	// Add authentication header
	auth := base64.StdEncoding.EncodeToString([]byte("riot:" + c.credentials.Password))
	headers := http.Header{}
	headers.Set("Authorization", "Basic "+auth)

	conn, _, err := dialer.Dial(wsURL, headers)
	if err != nil {
		return fmt.Errorf("failed to establish WebSocket connection: %w", err)
	}

	if c.config.Debug {
		c.logger.Debug("websocket", "WebSocket connection established successfully")
	}

	c.wsLock.Lock()
	c.wsConn = conn
	c.wsLock.Unlock()

	// Start listening for messages
	go c.listenForEvents()

	return nil
}

func (c *Client) sendWebSocketMessage(message interface{}) error {
	c.wsLock.RLock()
	defer c.wsLock.RUnlock()

	if c.wsConn == nil {
		return fmt.Errorf("WebSocket connection not established")
	}

	if c.config.Debug {
		c.logger.Debug("websocket", "Sending WebSocket message: %+v", message)
	}

	return c.wsConn.WriteJSON(message)
}

func (c *Client) listenForEvents() {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Error("websocket", "WebSocket listener panic: %v", r)
		}
	}()

	for {
		select {
		case <-c.done:
			return
		default:
			var message []interface{}
			if err := c.wsConn.ReadJSON(&message); err != nil {
				c.logger.Error("websocket", "Failed to read WebSocket message: %v", err)
				return
			}

			if len(message) > 0 {
				// First, pass the raw message to OnJsonApiEvent handlers
				c.eventMux.RLock()
				handlers := c.handlers["OnJsonApiEvent"]
				c.eventMux.RUnlock()

				for _, handler := range handlers {
					go handler(&Event{
						EventType: "WebSocketMessage",
						URI:       "OnJsonApiEvent",
						Data:      message,
					})
				}

				// Then process specific events if it's an event message
				if opcode, ok := message[0].(float64); ok {
					switch opcode {
					case 8: // EVENT
						c.handleEvent(message)
					}
				}
			}
		}
	}
}

func (c *Client) handleEvent(message []interface{}) {
	if len(message) < 3 {
		return
	}

	eventName, ok := message[1].(string)
	if !ok {
		return
	}

	eventData, ok := message[2].(map[string]interface{})
	if !ok {
		return
	}

	event := &Event{
		EventType: eventData["eventType"].(string),
		URI:       eventData["uri"].(string),
		Data:      eventData["data"],
	}

	// Get handlers for the event
	c.eventMux.RLock()
	var handlers []EventHandler

	// If this is an OnJsonApiEvent, we want to use the URI from the event data
	if eventName == "OnJsonApiEvent" {
		// Get handlers for the specific URI
		handlers = append(handlers, c.handlers[event.URI]...)
		// Get handlers for the root path (which catches all events)
		handlers = append(handlers, c.handlers["/"]...)
	} else {
		// Otherwise use the event name
		handlers = append(handlers, c.handlers[eventName]...)
	}
	c.eventMux.RUnlock()

	// Execute all handlers
	for _, handler := range handlers {
		go handler(event)
	}
}

// findCredentials attempts to find LCU connection credentials
func findCredentials(config *Config) (*Credentials, error) {
	// Try lockfile method first
	if creds, err := findCredentialsFromLockfile(config); err == nil {
		return creds, nil
	}

	// Try process method
	if creds, err := findCredentialsFromProcess(config); err == nil {
		return creds, nil
	}

	if config.AwaitConnection {
		return waitForCredentials(config)
	}

	return nil, fmt.Errorf("no running LCU instance found")
}

func findCredentialsFromLockfile(config *Config) (*Credentials, error) {
	var possiblePaths []string

	// If a custom path is provided, use it first
	if config.LeaguePath != "" {
		possiblePaths = append(possiblePaths, filepath.Join(config.LeaguePath, "lockfile"))
	}

	// Add platform-specific default paths
	switch runtime.GOOS {
	case "windows":
		// Try common drive letters
		for _, drive := range []string{"C", "D", "E", "F", "G"} {
			possiblePaths = append(possiblePaths, filepath.Join(drive+":", "Riot Games", "League of Legends", "lockfile"))
		}
	case "darwin":
		possiblePaths = append(possiblePaths, "/Applications/League of Legends.app/Contents/LoL/lockfile")
	case "linux":
		// Check if we're in WSL2 by looking for the Windows lockfile
		for _, drive := range []string{"c", "d", "e", "f", "g"} {
			possiblePaths = append(possiblePaths, filepath.Join("/mnt", drive, "Riot Games", "League of Legends", "lockfile"))
		}
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	// Try each possible path
	for _, path := range possiblePaths {
		if config.Debug {
			config.Logger.Debug("lockfile", "Trying lockfile path: %s", path)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			continue // Try next path
		}

		parts := strings.Split(string(data), ":")
		if len(parts) != 5 {
			continue // Invalid format, try next path
		}

		port, err := strconv.Atoi(parts[2])
		if err != nil {
			continue // Invalid port, try next path
		}

		if config.Debug {
			config.Logger.Debug("lockfile", "Found valid lockfile at: %s", path)
		}

		return &Credentials{
			Port:     port,
			Password: parts[3],
			Protocol: parts[4],
		}, nil
	}

	return nil, fmt.Errorf("no valid lockfile found in any of the possible locations")
}

func findCredentialsFromProcess(config *Config) (*Credentials, error) {
	var cmd *exec.Cmd
	var processPath string

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("wmic", "PROCESS", "WHERE", "name='LeagueClientUx.exe'", "GET", "commandline")
	case "darwin":
		cmd = exec.Command("ps", "-A", "-o", "command", "|", "grep", "LeagueClientUx")
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Extract the process path from the output
	outputStr := string(output)
	if runtime.GOOS == "windows" {
		// For Windows, the path is in the commandline output
		pathRegex := regexp.MustCompile(`"([^"]+\\LeagueClientUx\.exe)"`)
		if matches := pathRegex.FindStringSubmatch(outputStr); len(matches) > 1 {
			processPath = matches[1]
		}
	} else if runtime.GOOS == "darwin" {
		// For macOS, the path is in the ps output
		pathRegex := regexp.MustCompile(`/Applications/League of Legends\.app/Contents/LoL/LeagueClientUx`)
		if matches := pathRegex.FindStringSubmatch(outputStr); len(matches) > 0 {
			processPath = matches[0]
		}
	}

	// If we found the process path, update the config's LeaguePath
	if processPath != "" {
		// Get the directory containing LeagueClientUx.exe
		leagueDir := filepath.Dir(processPath)
		if config.Debug {
			config.Logger.Debug("process", "Found League installation at: %s", leagueDir)
		}
		config.LeaguePath = leagueDir
	}

	return parseProcessOutput(outputStr)
}

func parseProcessOutput(output string) (*Credentials, error) {
	portRegex := regexp.MustCompile(`--app-port=(\d+)`)
	passwordRegex := regexp.MustCompile(`--remoting-auth-token=([\w-]+)`)

	portMatch := portRegex.FindStringSubmatch(output)
	passwordMatch := passwordRegex.FindStringSubmatch(output)

	if len(portMatch) < 2 || len(passwordMatch) < 2 {
		return nil, fmt.Errorf("failed to extract credentials from process")
	}

	port, err := strconv.Atoi(portMatch[1])
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}

	return &Credentials{
		Port:     port,
		Password: passwordMatch[1],
		Protocol: "https",
	}, nil
}

// checkLCUHealth verifies if the LCU API is ready to accept connections
func checkLCUHealth(creds *Credentials, timeout time.Duration, logger Logger) bool {
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	url := fmt.Sprintf("https://127.0.0.1:%d/lol-summoner/v1/current-summoner", creds.Port)
	logger.Debug("health", "Attempting health check at %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Debug("health", "Failed to create request: %v", err)
		return false
	}

	auth := base64.StdEncoding.EncodeToString([]byte("riot:" + creds.Password))
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := client.Do(req)
	if err != nil {
		logger.Debug("health", "Health check request failed: %v", err)
		return false
	}
	resp.Body.Close()

	success := resp.StatusCode == http.StatusOK
	logger.Debug("health", "Health check response status: %d", resp.StatusCode)
	return success
}

func waitForCredentials(config *Config) (*Credentials, error) {
	ticker := time.NewTicker(config.PollInterval)
	defer ticker.Stop()

	logger := config.Logger
	logger.Debug("connection", "Starting to wait for LCU credentials...")

	for range ticker.C {
		creds, err := findCredentialsFromProcess(config)
		if err != nil {
			logger.Debug("connection", "Failed to find credentials: %v", err)
			continue
		}

		logger.Debug("connection", "Found credentials on port %d, checking health...", creds.Port)
		if checkLCUHealth(creds, config.Timeout, logger) {
			logger.Debug("connection", "Health check passed, LCU is ready")
			return creds, nil
		}
		logger.Debug("connection", "Health check failed, continuing to wait...")
	}
	return nil, fmt.Errorf("failed to find credentials after waiting")
}
