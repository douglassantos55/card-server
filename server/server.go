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

	server   *http.Server
	upgrader websocket.Upgrader
}

func NewServer() *Server {
	return &Server{
		Status: make(chan int),

		server:   &http.Server{},
		upgrader: websocket.Upgrader{},
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
	socket, err := s.upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("Could not upgrade connection")
		return
	}

	go func() {
		socket.WriteMessage(websocket.TextMessage, []byte("Welcome"))

	outer:
		for {
			select {
			case status := <-s.Status:
				if status == 3 {
					socket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
					break outer
				}
			}
		}

		s.Status <- 4
	}()
}
