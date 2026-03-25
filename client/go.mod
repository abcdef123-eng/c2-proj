module github.com/execute-assembly/c2-proj/newClient

go 1.26.1

require (
	github.com/execute-assembly/c2-proj/modules v0.0.0-00010101000000-000000000000
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/peterh/liner v1.2.2
	google.golang.org/grpc v1.79.3
)

replace github.com/execute-assembly/c2-proj/modules => ../modules

require (
	github.com/mattn/go-runewidth v0.0.3 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
