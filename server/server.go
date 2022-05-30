package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Server struct {
	Status chan int

	server     *http.Server
	dispatcher *Dispatcher
	upgrader   websocket.Upgrader
}

func NewServer(dispatcher *Dispatcher) *Server {
	return &Server{
		Status: make(chan int),

		dispatcher: dispatcher,
		server:     &http.Server{},
		upgrader:   websocket.Upgrader{},
	}
}

func (s *Server) Close() {
	s.server.Shutdown(context.Background())
}

func (s *Server) Listen(addr string) {
	s.server.Addr = addr
	s.server.Handler = http.HandlerFunc(s.handleConnection)

	err := s.server.ListenAndServe()

	if err == http.ErrServerClosed {
		s.Status <- 3
	}
}

func (s *Server) ListenQuietly(addr string) {
	go s.Listen(addr)
	time.Sleep(time.Millisecond)
}

func (s *Server) handleConnection(w http.ResponseWriter, r *http.Request) {
	s.upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	socket, err := s.upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("Could not upgrade connection")
		return
	}

	player := NewPlayer(socket)
	player.Send(Response{
		Type: Welcome,
	})

	go func() {
		defer player.Close()

		for {
			select {
			case event := <-player.Incoming:
				event.Player = player
				s.dispatcher.Dispatch <- event
			}
		}
	}()
}
