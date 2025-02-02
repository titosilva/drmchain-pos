package cmdserver

import (
	"log"
	"net/http"
)

type CommandsServer struct {
	httpMux *http.ServeMux
}

func NewCommandsServer() *CommandsServer {
	s := CommandsServer{
		httpMux: http.NewServeMux(),
	}

	s.ConfigureEndpoints()

	return &s
}

func (s *CommandsServer) ConfigureEndpoints() {
	s.httpMux.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

func (s *CommandsServer) Start() {
	port := ":2502" // TODO: move to config
	log.Println("Starting http server at port " + port)

	go http.ListenAndServe(port, s.httpMux)
}
