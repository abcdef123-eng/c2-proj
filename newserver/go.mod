module github.com/execute-assembly/c2-proj/newserver

go 1.26.1

require (
	github.com/abcdef123-eng/c2-proj/modules v0.0.0
	google.golang.org/grpc v1.79.3
)

replace github.com/abcdef123-eng/c2-proj/modules => ../modules

require (
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)
