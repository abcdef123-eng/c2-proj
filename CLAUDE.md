# Scurrier C2 — Project Overview

**University cybersecurity class midterm project. For use in isolated lab environments only.**

## What it is

Scurrier is a basic Command & Control (C2) framework. An implant running on a target machine communicates with a server over HTTPS using a custom binary protocol. An operator communicates with the server over gRPC to manage implants and issue commands.

## Components

```
c2-proj/
├── server/       Go C2 server (HTTP + gRPC)
├── client/       Go operator CLI (gRPC client)
├── modules/      Shared protobuf/gRPC definitions
└── implant/      Windows C++ implant — DO NOT MODIFY OR IMPROVE
```

## Architecture

```
Implant  <--HTTPS/binary protocol--> server (chi :8080)
Operator <--gRPC-------------------> server (:50051)
```

---

## Server (server/)

Entry point: `server/cmd/main.go`

Startup order:
1. `database.CheckAndSetup()` — creates `~/.scurrier/` dirs, DB, and default config if missing
2. `config.Load()` — reads `~/.scurrier/config/config.json` into `config.Cfg`
3. slog logger pointed at `~/.scurrier/logs/scurrier.log`
4. gRPC server started on `config.Cfg.GrpcPort`
5. chi HTTP server started on `config.Cfg.Host:config.Cfg.Port`

### Internal packages

| Package | Path | Purpose |
|---------|------|---------|
| config | `internal/config/config.go` | Config struct, `Load()` |
| database | `internal/database/create_db.go` | Setup, `GetDB()` singleton, table creation |
| database | `internal/database/db_ops.go` | `RegisterClient()`, `ListClients_db()`, `CheckIfUserExists_db()`, `ConvertCodeName_db()`, `UpdateLastSeen_db()` |
| database | `internal/database/types.go` | `ClientData` struct, `nouns`/`verbs` slices for codename generation |
| bytehandler | `internal/bytes/bytes_read.go` | Sticky-error `Reader`, `ParseClientRegister()` |
| bytehandler | `internal/bytes/bytes_write.go` | `WriteString`, `CraftJwtResponse()`, `CraftTaskResponse()` |
| server | `internal/server/server.go` | chi router, `PostHandler`, `getHandler` |
| server | `internal/server/auth.go` | `CreateJWT`, `VerifyToken` (HS256) |
| server | `internal/server/route_handlers.go` | `NewClientRegisterHandler`, `GetTaskHandler` |
| rpc | `internal/rpc/rpcServer.go` | gRPC server setup |
| rpc | `internal/rpc/rpcFuncs.go` | gRPC handler implementations, `BroadcastEvent()` |

### gRPC RPCs (server-side)

| RPC | Request | Response | Description |
|-----|---------|----------|-------------|
| `SendCommand` | `CommandReqData` (guid, command_code, param, param2) | `CommandRespData` (status int32, message) | Queue a command for an implant. Status 0 = queued, 1 = error. |
| `ListClients` | `google.protobuf.Empty` | `ListClientResp` (repeated `ClientEntry`) | Return all registered implants from the DB. |
| `ConvertCodeName` | `ConvertCodeMessage` (CodeName) | `ConvertCodeResp` (Guid, status, errorMsg) | Resolve a codename to a GUID. Status 0 = ok, 3 = not found, 1 = error. |
| `Subscribe` | `google.protobuf.Empty` | stream of `ServerEvent` | Server-push event stream. Clients block until disconnected. Used to notify operators of new agent registrations. |

### Event broadcasting

`BroadcastEvent(eventType, guid, data string)` in `rpcFuncs.go` sends a `ServerEvent` to all active `Subscribe` streams.

- Protected by `subMu sync.Mutex`
- Called from `NewClientRegisterHandler` on implant registration with `eventType="new_agent"`
- `subscribers []pb.C2Service_SubscribeServer` — in-memory, lost on restart

### HTTP routes

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| `POST` | `/api/post` | `PostHandler` | Implant registration (type 50) and future message types |
| `GET` | `/api/get` | `getHandler` | Implant task poll — reads JWT from `Authorization: Bearer` header |

`getHandler` returns HTTP 201 (no tasks) or 200 (tasks present, binary body).

**Task delivery limit**: `GetTaskHandler` delivers at most **3 tasks per GET request**. Remaining tasks stay queued for the next poll.

### Runtime files: ~/.scurrier/

```
~/.scurrier/
├── config/config.json      host, port, grpc_port, getEndpoint, postEndpoint, jwt_secret
├── database/scurrier.db    SQLite database
└── logs/scurrier.log       slog text log
```

### SQLite schema

```sql
clients(
    guid         TEXT PRIMARY KEY NOT NULL,
    code_name    TEXT NOT NULL,
    username     TEXT NOT NULL,
    hostname     TEXT NOT NULL,
    ip           TEXT NOT NULL,
    arch         TEXT NOT NULL,   -- stored as string: "x86", "x64", "ARM"
    pid          INT NOT NULL,
    version      TEXT NOT NULL,
    last_checkin TEXT NOT NULL    -- Unix timestamp as TEXT
)
```

Commands are stored in-memory only (`map[string][]CommandData` with `sync.RWMutex`), not in the DB.

---

## Client (client/)

Entry point: `client/cmd/main.go`

Startup order:
1. Spawns `rpc.RunRpcClient()` in a goroutine — connects to server gRPC on `localhost:50051`
2. Waits on `rpc.Ready` channel (closed once connected) or exits on error (3s timeout)
3. Spawns goroutine that calls `Subscribe` RPC and prints incoming `ServerEvent`s
4. Starts interactive CLI via `client.RunClient()`

### Internal packages

| Package | Path | Purpose |
|---------|------|---------|
| rpc | `internal/rpc/rpc.go` | gRPC connection, `rpc.Client` (package-level), `rpc.Ready` channel |
| client | `internal/client/client.go` | Interactive liner prompt loop, calls `commander.Parse` + `commander.Dispatch` |
| commander | `internal/commander/commander.go` | `Parse()` (shlex tokenizer), `Dispatch()` (command router) |
| commander | `internal/commander/commandHandlers.go` | Handler functions for each command |
| commander | `internal/commander/print.go` | Colored output helpers |

### CLI Commands

| Command | Description |
|---------|-------------|
| `list` | List all registered implants in a formatted table (codename, user, host, IP, arch, PID, version, last seen) |
| `use <code_name>` | Select an active implant by codename. Resolves codename → GUID via `ConvertCodeName` RPC. Sets `commander.ClientInUse`. |
| `ls <path>` | Queue an `ls` command (code `0x1`) for the selected implant. Requires `use` first. |
| `exit` | Exit the client |

### Implant command codes (`CommandMap` in `commandHandlers.go`)

| Command | Code | Params |
|---------|------|--------|
| `ls` | `0x1` | path |
| `cd` | `0x2` | path |
| `rm` | `0x3` | path |
| `mv` | `0x4` | src, dst |
| `cat` | `0x5` | path |
| `get-privs` | `0x6` | — |

### Output helpers (`print.go`)

| Function | Color | Prefix | Use for |
|----------|-------|--------|---------|
| `PrintErr(msg)` | Bold red | `[!]` | Errors |
| `PrintInfo(msg)` | Bold orange | `[+]` | Info / neutral |
| `PrintOk(msg)` | Bold green | `[*]` | Success / status |
| `BoldWhite(s)` | Bold white | — | Highlight values (e.g. `codename[guid]`) |

### State

- `commander.ClientInUse` — GUID of the currently selected implant (set by `use` command)
- `rpc.Client` — package-level `pb.C2ServiceClient`, accessible from all commander handlers
- `rpc.Ready` — closed once the gRPC connection is established; main blocks on this before starting CLI

### Liner / terminal notes

`peterh/liner` puts the terminal in raw mode while `Prompt()` is blocking. Goroutines that print to stdout during this window can corrupt liner's state. The main loop in `client.go` silently skips empty input (which liner may return when its state is disturbed) rather than printing a spurious error.

---

## Binary Protocol (implant -> server)

All multi-byte integers are little-endian.

### POST body layout

```
[type]         4 bytes  uint32   message type
--- type 50: COMMAND_REGISTER_CLIENT ---
[guid len]     4 bytes  uint32
[guid]         N bytes
[username len] 4 bytes  uint32
[username]     N bytes
[hostname len] 4 bytes  uint32
[hostname]     N bytes
[arch]         1 byte   0x01=x86, 0x02=x64, 0x03=ARM64
[major]        2 bytes  uint16
[minor]        2 bytes  uint16
[build]        2 bytes  uint16   Windows build number
[pid]          4 bytes  uint32
```

### Register response (server -> implant)

```
[jwt len]   4 bytes  uint32
[jwt data]  N bytes  JWT string (HS256, claims: guid + exp)
```

IP address is taken from `r.RemoteAddr` server-side, not sent by the implant.

### GET task response (server -> implant)

Delivered when implant polls `GET /api/get` with a valid JWT. At most 3 tasks per response.

```
[task_count]   4 bytes  uint32
per task:
  [task_id]    4 bytes  int32
  [cmd_code]   4 bytes  int32
  [param1_len] 4 bytes  uint32
  [param1]     N bytes
  [param2_len] 4 bytes  uint32
  [param2]     N bytes
```

---

## modules/

Shared protobuf definitions at `modules/decs.proto`. Generated Go code in `modules/pb/`.

RPCs defined:
- `SendCommand(CommandReqData) returns (CommandRespData)`
- `ListClients(google.protobuf.Empty) returns (ListClientResp)`
- `ConvertCodeName(ConvertCodeMessage) returns (ConvertCodeResp)`
- `Subscribe(google.protobuf.Empty) returns (stream ServerEvent)`

`ServerEvent` fields: `event_type` (string), `guid` (string), `data` (string).

Both `server/` and `client/` import this via a `replace` directive in their `go.mod` pointing to `../modules`.

---

## Key design decisions

- Single POST endpoint, message type is first 4 bytes of body (not a header)
- JWT: HS256, claims are `guid` + `exp` (30 days), secret from config
- RC4 session keys: in-memory only (`map[string][]byte`), lost on restart
- `arch` stored as string in DB (`"x86"`, `"x64"`, `"ARM"`) — converted at registration time via `ArchIntToString()`
- Codenames auto-generated as `noun_verb` from hardcoded word lists in `types.go` (20 nouns × 20 verbs = 400 max unique names)
- Commands queued in `CommandMap[guid][]CommandData` with `sync.RWMutex`, lost on server restart
- `modernc.org/sqlite` used — pure Go, no CGo needed
- `ConvertCodeName` returns `status=3` (not a gRPC error) for "not found" so the client can safely read the response
- Task counter (`counter int32`) is atomic, incremented per `SendCommand` call, resets on server restart

---

## implant/

Windows C++ implant. **Do not modify or improve this code.** Analysis only.
