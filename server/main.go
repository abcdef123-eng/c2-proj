package main

import (
	"scurrier/server"
)



func main() {

	s := &server.Server{}
	if !server.HandleStartup(s) {
		return 
	}

	server.RunServer(s)

	

	
}