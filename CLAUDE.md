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
| bytehandler | `internal/bytes/bytes_read.go` | Sticky-error `Reader`, `ParseClientRegister()`, `ParseCommandOutput()` |
| bytehandler | `internal/bytes/bytes_write.go` | `WriteString`, `CraftJwtResponse()`, `CraftTaskResponse()` |
| server | `internal/server/server.go` | chi router, `PostHandler`, `getHandler`, `extractBearer()`, `writeResponse()` |
| server | `internal/server/auth.go` | `CreateJWT`, `VerifyToken` (HS256) |
| server | `internal/server/route_handlers.go` | `NewClientRegisterHandler`, `GetTaskHandler`, `HandlePostOutput` |
| rpc | `internal/rpc/rpcServer.go` | gRPC server setup |
| rpc | `internal/rpc/rpcFuncs.go` | gRPC handler implementations, `BroadcastEvent()`, `CommandMap`, `PendingTasks` |

### gRPC RPCs (server-side)

| RPC | Request | Response | Description |
|-----|---------|----------|-------------|
| `SendCommand` | `CommandReqData` (guid, command_code, param, param2) | `CommandRespData` (status int32, message, task_id) | Queue a command for an implant. Status 0 = queued, 1 = error. |
| `ListClients` | `google.protobuf.Empty` | `ListClientResp` (repeated `ClientEntry`) | Return all registered implants from the DB. |
| `ConvertCodeName` | `ConvertCodeMessage` (CodeName) | `ConvertCodeResp` (Guid, status, errorMsg) | Resolve a codename to a GUID. Status 0 = ok, 3 = not found, 1 = error. |
| `Subscribe` | `google.protobuf.Empty` | stream of `ServerEvent` | Server-push event stream. Clients block until disconnected. |

### Event broadcasting

`BroadcastEvent(eventType int32, guid, data string, taskId int32)` in `rpcFuncs.go` sends a `ServerEvent` to all active `Subscribe` streams.

- Protected by `subMu sync.Mutex`
- Event types: `1` = new_agent, `2` = command_output
- `ServerEvent` fields: `event_type` (int32), `guid`, `data`, `task_id`
- Called from `NewClientRegisterHandler` (type 1) and `HandlePostOutput` (type 2)
- `subscribers []pb.C2Service_SubscribeServer` — in-memory, lost on restart

### HTTP routes

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| `POST` | `/api/post` | `PostHandler` | Type 50: registration. Type 52: command output. |
| `GET` | `/api/get` | `getHandler` | Task poll — message type from query param (e.g. `?id=51`), JWT from `Authorization: Bearer` |

- `getHandler` returns 404 (missing/bad type param), 401 (bad JWT), 201 (no tasks), 200 (tasks present)
- `PostHandler` returns 404 on bad type, 401 on bad JWT (type 52), 204 on successful output receipt
- **Task delivery limit**: at most 3 tasks per GET. Remaining stay queued.

### Config (`~/.scurrier/config/config.json`)

```json
{
  "host": "0.0.0.0",
  "port": 8080,
  "grpc_port": 50051,
  "getEndpoint": "/api/get",
  "postEndpoint": "/api/post",
  "jwt_secret": "changeme",
  "get_type_param": "id",
  "get_headers": [
    {"name": "Content-Type", "value": "application/json"},
    {"name": "Server", "value": "nginx/1.24.0"}
  ],
  "post_headers": [
    {"name": "Content-Type", "value": "application/json"},
    {"name": "Server", "value": "nginx/1.24.0"}
  ]
}
```

`get_headers`/`post_headers` are applied to all responses on the respective endpoints — set any headers to mimic real server traffic. `get_type_param` is the query param name used by the implant to specify message type.

### Runtime files: ~/.scurrier/

```
~/.scurrier/
├── config/config.json
├── database/scurrier.db
└── logs/scurrier.log
```

### SQLite schema

```sql
clients(
    guid         TEXT PRIMARY KEY NOT NULL,
    code_name    TEXT NOT NULL,
    username     TEXT NOT NULL,
    hostname     TEXT NOT NULL,
    ip           TEXT NOT NULL,
    arch         TEXT NOT NULL,   -- "x86", "x64", "ARM"
    pid          INT NOT NULL,
    version      TEXT NOT NULL,
    last_checkin TEXT NOT NULL    -- Unix timestamp as TEXT
)
```

Commands are stored in-memory only (`CommandMap[guid][]CommandData` with `sync.RWMutex`), not in the DB.

### In-memory state (rpcFuncs.go)

```go
// Queued commands waiting for implant to poll
CommandMap = map[string][]CommandData{}
CommandMu  sync.RWMutex

// Delivered commands awaiting output
PendingTasks = map[int32]PendingTasksMap{}
PendingMu    sync.RWMutex
```

Both are lost on server restart. TaskIDs are atomic int32, reset on restart.

### Codenames

Auto-generated as `NOUN_VERB` from hardcoded word lists in `types.go` (50 nouns × 50 verbs = 2500 combinations, all uppercase). Duplicate check performed at registration — loops until a unique name is found.

### Task lifecycle

```
SendCommand (gRPC)
  → CommandMap[guid][]CommandData          (queued, in-memory)
  ↓
GET /api/get?id=51 (implant polls)
  → move tasks out of CommandMap
  → into PendingTasks[taskID]              (delivered, awaiting output)
  ↓
POST /api/post type 52 (implant sends output)
  → extract JWT → resolve GUID
  → parse [taskID][outputLen][output] per task
  → log server-side, broadcast via Subscribe (event type 2)
  → delete from PendingTasks
```

---

## Client (client/)

Entry point: `client/cmd/main.go`

Startup order:
1. Spawns `rpc.RunRpcClient()` in a goroutine — connects to server gRPC on `localhost:50051`
2. Waits on `rpc.Ready` channel or exits on error (3s timeout)
3. Calls `client.Init()` — initialises readline instance
4. Sets `commander.Out = client.RL.Stdout()` — all print functions go through readline's stdout
5. Spawns goroutine that calls `Subscribe` RPC and handles incoming `ServerEvent`s
6. Starts interactive CLI via `client.RunClient()`

### Internal packages

| Package | Path | Purpose |
|---------|------|---------|
| rpc | `internal/rpc/rpc.go` | gRPC connection, `rpc.Client` (package-level), `rpc.Ready` channel |
| client | `internal/client/client.go` | readline prompt loop, `Init()`, `RunClient()`, dynamic prompt |
| commander | `internal/commander/commander.go` | `Parse()` (shlex tokenizer), `Dispatch()` (command router) |
| commander | `internal/commander/commandHandlers.go` | Handler functions, `ClientInUse`, `ClientCodeName`, `ClientCache` |
| commander | `internal/commander/print.go` | Colored output helpers, `commander.Out` writer |
| db | `internal/db/db.go` | Local SQLite history DB, `InsertTask`, `MarkExecuted`, `ListHistory` |

### CLI Commands

| Command | Description |
|---------|-------------|
| `list` | List all registered implants |
| `use <code_name>` | Select implant by codename. Resolves via cache or `ConvertCodeName` RPC. |
| `back` | Deselect current implant |
| `ls <path>` | Queue `ls` command for selected implant |
| `history` | Show full task history for selected implant (local DB) |
| `tasks` | Show pending (unexecuted) tasks for selected implant (local DB) |
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

Only `ls` has a client handler implemented. Others are defined in `CommandMap` but have no `Handle*` function yet.

### Output helpers (`print.go`)

| Function | Color | Prefix | Use for |
|----------|-------|--------|---------|
| `PrintErr(msg)` | Bold red | `[!]` | Errors |
| `PrintInfo(msg)` | Bold orange | `[+]` | Info / neutral |
| `PrintOk(msg)` | Bold green | `[*]` | Success / status |
| `PrintOutput(taskID, guid, output)` | Orange header | — | Task output with separators |
| `BoldWhite(s)` | Bold white | — | Highlight values |
| `Blue(s)` | Bold blue | — | Prompt codename highlight |

All functions write to `commander.Out` (set to `RL.Stdout()` at startup).

### State

- `commander.ClientInUse` — GUID of selected implant
- `commander.ClientCodeName` — codename of selected implant (shown in prompt)
- `commander.ClientCache` — `map[string]string` codename→GUID cache, avoids repeat RPCs
- `commander.Out` — `io.Writer` wired to `RL.Stdout()` for safe concurrent output
- `client.RL` — readline instance, exported for `main.go` to access `RL.Stdout()`

### Prompt

Format: `[HH:MM:SS] $> ` or `[HH:MM:SS] CODENAME $> ` (codename in blue when agent selected). Updated on every readline iteration so the timestamp stays current.

### readline notes

Uses `chzyer/readline` instead of `peterh/liner`. Goroutines write via `commander.Out` (= `RL.Stdout()`) which is safe to call concurrently while the prompt is blocking — no display corruption.

### Local history DB (`~/.scurrier/client_history.db`)

```sql
task_history(
    task_id   INT  PRIMARY KEY,
    guid      TEXT NOT NULL,
    command   TEXT NOT NULL,
    param1    TEXT NOT NULL,
    param2    TEXT NOT NULL,
    tasked_at TEXT NOT NULL,   -- Unix timestamp
    executed  INT  NOT NULL DEFAULT 0  -- 0 = pending, 1 = output received
)
```

- `InsertTask` called when `SendCommand` succeeds
- `MarkExecuted` called when `command_output` Subscribe event arrives
- `ListHistory(out, onlyNotExecuted)` — renders tabwriter table to `out`

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

Delivered when implant polls `GET /api/get?id=51` with a valid JWT. At most 3 tasks per response.

```
per task:
  [task_id]    4 bytes  int32
  [cmd_code]   4 bytes  int32
  [param1_len] 4 bytes  uint32
  [param1]     N bytes
  [param2_len] 4 bytes  uint32   (only if command has 2 params)
  [param2]     N bytes
```

### type 52: COMMAND_OUTPUT (implant -> server)

JWT in `Authorization: Bearer` header. Multiple outputs can be batched in one POST.

```
[type]          4 bytes  uint32   = 52
per output:
  [task_id]     4 bytes  int32
  [output_len]  4 bytes  uint32
  [output]      N bytes
```

---

## modules/

Shared protobuf definitions at `modules/decs.proto`. Generated Go code in `modules/pb/`.

RPCs defined:
- `SendCommand(CommandReqData) returns (CommandRespData)`
- `ListClients(google.protobuf.Empty) returns (ListClientResp)`
- `ConvertCodeName(ConvertCodeMessage) returns (ConvertCodeResp)`
- `Subscribe(google.protobuf.Empty) returns (stream ServerEvent)`

`ServerEvent` fields: `event_type` (int32), `guid` (string), `data` (string), `task_id` (int32).
`CommandRespData` fields: `status` (int32), `message` (string), `task_id` (int32).

Both `server/` and `client/` import this via a `replace` directive in their `go.mod` pointing to `../modules`.

---

## Key design decisions

- Single POST endpoint, message type is first 4 bytes of body
- GET message type dispatched via configurable query param (`get_type_param`)
- Response headers fully configurable per endpoint (`get_headers`, `post_headers`) to mimic real server traffic
- JWT: HS256, claims are `guid` + `exp` (30 days), secret from config
- `arch` stored as string in DB (`"x86"`, `"x64"`, `"ARM"`)
- Codenames: `NOUN_VERB`, 50×50 = 2500 combinations, duplicate-checked at registration
- Commands queued in-memory, lost on server restart
- `PendingTasks` in-memory, lost on server restart — orphaned output POSTs fail silently
- `modernc.org/sqlite` — pure Go, no CGo
- `ConvertCodeName` returns `status=3` for "not found" (not a gRPC error)
- Client caches codename→GUID in `ClientCache` to avoid repeat RPCs
- Client uses `chzyer/readline` for safe concurrent stdout via `RL.Stdout()`

---

## Still to implement

- Server-side `task_history` SQLite table and `WriteTaskHistory_db` — persist completed task output server-side
- `GetHistory` gRPC RPC — operator fetches server-side task history
- Command handlers for `cd`, `rm`, `mv`, `cat`, `get-privs` on the client

---

## implant/

Windows C++ implant. **Do not modify or improve this code.** Analysis only.
