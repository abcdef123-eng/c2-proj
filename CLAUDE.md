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
| database | `internal/database/db_ops.go` | `RegisterClient()` |
| bytehandler | `internal/bytes/bytes_read.go` | Sticky-error `Reader`, `ParseClientRegister()` |
| bytehandler | `internal/bytes/bytes_write.go` | `WriteString`, `CraftJwtResponse()` |
| server | `internal/server/server.go` | chi router, `PostHandler` |
| server | `internal/server/auth.go` | `CreateJWT`, `VerifyToken` (HS256) |
| server | `internal/server/route_handlers.go` | `NewClientRegisterHandler` and other handlers |
| rpc | `internal/rpc/rpcServer.go` | gRPC server setup |
| rpc | `internal/rpc/rpcFuncs.go` | gRPC handler implementations |

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
    arch         INT NOT NULL,
    pid          INT NOT NULL,
    version      TEXT NOT NULL,
    last_checkin TEXT NOT NULL
)
```

Commands are stored in-memory only (`map[string][]CommandData` with `sync.RWMutex`), not in the DB.

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

## Key design decisions

- Single POST endpoint, message type is first 4 bytes of body (not a header)
- JWT: HS256, claims are `guid` + `exp` (30 days), secret from config
- RC4 session keys: in-memory only (`map[string][]byte`), lost on restart
- `modernc.org/sqlite` used — pure Go, no CGo needed
- `modules/` uses a `replace` directive in `go.mod` pointing to `../modules` (local monorepo pattern)

## modules/

Shared protobuf definitions at `modules/decs.proto`. Generated Go code in `modules/pb/`.

Current RPC: `SendCommand(CommandReqData) returns (CommandRespData)`

## implant/

Windows C++ implant. **Do not modify or improve this code.** Analysis only.
