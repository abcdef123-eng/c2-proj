package main

import (
	"fmt"
	"log"

	"github.com/execute-assembly/c2-proj/newserver/internal/config"
	"github.com/execute-assembly/c2-proj/newserver/internal/database"
	"github.com/execute-assembly/c2-proj/newserver/internal/rpc"
	"github.com/execute-assembly/c2-proj/newserver/internal/server"
)

func main() {

	err := database.CheckAndSetup()
	if err := config.Load(); err != nil {
		log.Fatalf("[!] Failed to load config: %v", err)
	}
	if err != nil {
		fmt.Println(err)
		return
	}

	errCh := make(chan error, 1)

	go func() {
		errCh <- rpc.RunRpcServer()
	}()
	if err := <-errCh; err != nil {
		log.Fatalf("[!] gRPC Failed to Start: %v", err)
	}
	fmt.Println("[+] gRPC Server Started..")

	server.StartServer()

	return

}
