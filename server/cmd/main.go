package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/execute-assembly/c2-proj/server/internal/config"
	"github.com/execute-assembly/c2-proj/server/internal/database"
	"github.com/execute-assembly/c2-proj/server/internal/rpc"
	"github.com/execute-assembly/c2-proj/server/internal/server"
)

func main() {

	if err := database.CheckAndSetup(); err != nil {
		log.Fatalf("[!] Failed to setup: %v", err)
	}
	if err := config.Load(); err != nil {
		log.Fatalf("[!] Failed to load config: %v", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("[!] Failed to get home dir: %v", err)
	}
	f, err := os.OpenFile(homeDir+"/.scurrier/logs/scurrier.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(f, nil))
	slog.SetDefault(logger)

	errCh := make(chan error, 1)

	go func() {
		errCh <- rpc.RunRpcServer()
	}()
	if err := <-errCh; err != nil {
		log.Fatalf("[!] gRPC Failed to Start: %v", err)
	}
	fmt.Println("[+] gRPC Server Started..")

	server.StartServer()
}
