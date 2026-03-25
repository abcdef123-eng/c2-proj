package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/execute-assembly/c2-proj/newserver/internal/config"

	"github.com/go-chi/chi/v5"
)

const (
	COMMAND_TYPE_REGISTER = 50
)

func StartServer() {

	r := chi.NewRouter()

	//r.Get(config.Cfg.GetEndpoint, GetHandler)
	r.Post(config.Cfg.PostEndpoint, PostHandler)

	// r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("Hello Chi!"))
	// })

	http.ListenAndServe(fmt.Sprintf("%s:%d", config.Cfg.Host, config.Cfg.Port), r)

}

func PostHandler(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed Reading POST Body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	reader := bytes.NewReader(body)
	var CommandType uint32
	binary.Read(reader, binary.LittleEndian, &CommandType)

	switch CommandType {
	case COMMAND_TYPE_REGISTER:
		Ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		JwtBytes, err := NewClientRegisterHandler(Ip, reader)
		if err != nil {
			http.Error(w, "Failed", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(JwtBytes)
	}

}
