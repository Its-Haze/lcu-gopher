# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`lcu-gopher` is a single-package Go library (`package lcu`, module `github.com/its-haze/lcu-gopher`) for talking to the League of Legends Client API. It does three things: find the running client's credentials, make authenticated HTTPS requests to it, and dispatch its WebSocket events to handlers.

Dependencies are `github.com/gorilla/websocket` and `github.com/shirou/gopsutil/v4`. Four source files, ~1600 lines. Credential discovery is tested; nothing else is.

## Commands

```powershell
go build ./...
go vet ./...
gofmt -w .
go run ./example/request        # one GET, prints raw JSON
go run ./example/subscribe      # AwaitConnection + Subscribe to summoner updates
go run ./example/gameflowphase  # SubscribeToGamePhase
```

`go test ./...` covers credential discovery only. Every example needs a running League Client; without one `NewClient` fails immediately (or blocks, if `AwaitConnection` is set).

Releases are plain git tags (`v0.0.0` through `v0.0.3`). Consumers pull the tagged version from the module proxy, not a local path, so a change here is invisible downstream until it is tagged and the consumer's `go.mod` is bumped.

## Architecture

### Connection lifecycle

`NewClient(config)` does more than allocate. It resolves credentials before returning, in this order:

1. `findCredentialsFromProcess` walks the process table through gopsutil, matches `LeagueClientUx.exe` by process name, and pulls `--app-port` and `--remoting-auth-token` out of the command line. It also back-fills `config.LeaguePath` from the exe's directory, so the lockfile method can find a non-default install.
2. `findCredentialsFromLockfile` reads `{leaguePath}/lockfile`, falling back to `{drive}:\Riot Games\League of Legends\lockfile` across drives C through G, parses the 5-field colon-separated format, takes fields 2 and 3 as port and password.
3. Both are health-checked against `/lol-summoner/v1/current-summoner` before being accepted, so a stale lockfile from a crashed client does not mask the working method.
4. If `AwaitConnection` is true, `waitForCredentials` retries **both** methods on every `PollInterval` tick.

Discovery must never shell out. It used to run `wmic`, which Windows 11 24H2 no longer ships, and that hung `AwaitConnection` callers forever with League running: the process scan was the only method the wait loop retried. `ProcessLister` on `Config` is the seam for faking the process table in tests.

`Connect()` then does an HTTP smoke test and dials the WebSocket. Two consequences that matter to callers:

- The blocking wait happens in `NewClient`, not `Connect`. Waiting for League to start means constructing the client late, not reconnecting an existing one.
- `Disconnect()` does `close(c.done)`, so a `Client` is single-use. Calling `Disconnect` twice panics, and `Connect` after `Disconnect` leaves the event listener dead. Reconnect means building a fresh `Client`.

There is no disconnect detection. Nothing notices League closing mid-session; `listenForEvents` just logs a read error and returns.

### Auth and TLS

Every request and the WebSocket handshake carry `Authorization: Basic base64("riot:" + password)`. Both use `InsecureSkipVerify: true` because the client serves a self-signed cert. That is correct here and not a bug to fix.

### Event dispatch (WAMP)

`Subscribe(endpoint, handler, eventTypes...)` wraps the handler in a type filter, then registers the wrapped handler under **two** keys, `endpoint` and `"OnJsonApiEvent"`, and sends WAMP subscribe frames (`[5, uri]`) for both. At least one event type is required; passing none is an error.

Inbound routing in `listenForEvents` and `handleEvent`:

- Every raw frame is first handed to the `"OnJsonApiEvent"` handler list as a synthetic event with `EventType: "WebSocketMessage"` and `Data` set to the raw `[]interface{}` message. The type filter in the `Subscribe` wrapper discards these, so a normal subscriber never sees them. A handler registered by other means would.
- Frames with opcode 8 go to `handleEvent`, which reads `eventType`, `uri` and `data` out of the payload and dispatches to handlers keyed on the event's `uri` plus handlers keyed on `"/"`.
- `"/"` is therefore the catch-all key, which is what `SubscribeToAll` registers under.

Handlers run in their own goroutine per event (`go handler(event)`), so they must be safe to run concurrently and out of order. `Event.Data` is `interface{}` holding decoded JSON, so consumers type-assert to `map[string]interface{}` and pick fields out by hand.

`handleEvent` type-asserts `eventData["eventType"]` and `["uri"]` without the two-value form, so a malformed opcode-8 frame panics the listener goroutine. The `recover()` in `listenForEvents` catches it and the listener stops.

### Logging

`Logger` is a three-method interface (`Info`/`Error`/`Debug`), each taking an `endpoint` string first. Two implementations ship: `defaultLogger` (stdout, drops debug lines unless `Debug` is set) and `EndpointLogger`, which additionally opens one append-mode file per endpoint under `LogDir`, named after the endpoint with `/` replaced by `_`.

`NewClient` swaps in an `EndpointLogger` whenever `LogDir` is non-empty, and sets `LogDir` to `"logs"` on its own if `Debug` is true and `LogDir` is empty. Setting `Debug` without `LogDir` writes files into the working directory.

A nil `Logger` panics: `waitForCredentials` and the lockfile/process finders call `config.Logger.Debug` directly with no nil check. Build configs from `DefaultConfig()` rather than a bare `&Config{}`.

### Files

- `client.go` (905 lines) is everything above: `Client`, `Config`, both loggers, credential discovery, HTTP, WebSocket, dispatch.
- `types.go` holds the LCU response structs (`Summoner`, `GameSession`, `ChampSelectSession`, `Lobby`, `Friend`, `RunePage`, ...) plus `GamePhase` and `EventType` constants and a queue-ID block. The queue-ID comments contradict themselves (`QueueNormalBlind = 430` is labelled "incorrect"); trust the LCU schema, not these.
- `helpers.go` has the typed convenience getters. Each one is the same shape: `Get`, check status, decode into a struct. `RankedStats` is declared here rather than in `types.go`.
- `errors.go` has five sentinel errors, of which only `ErrSummonerNotFound` is actually used.

## Conventions

Adding a typed endpoint helper means: struct in `types.go`, method in `helpers.go` following the existing get-check-decode shape, endpoint path inline in the method (there is no endpoint constant block). Wrap decode failures with `%w`.

Public API is documented with full godoc comments, including the parameter/return lists on the bigger functions. Match that if you touch an exported symbol.

## Downstream

This library is consumed by `../league-rpc/`. See `CLAUDE.local.md` for how, and for which quirks above that project already works around.

LCU endpoint reference: https://www.mingweisamuel.com/lcu-schema/tool/#/
