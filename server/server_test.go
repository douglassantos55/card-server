package server

import (
	"testing"
	"time"
)

func TestAcceptsConnections(t *testing.T) {
	server := NewServer()
	defer server.Close()

	server.ListenQuietly("0.0.0.0:8080")
	client := NewClient("0.0.0.0:8080")

	select {
	case <-client.Incoming:
	case <-time.After(time.Second):
		t.Error("Expected welcome from server")
	}
}

func TestClosesServer(t *testing.T) {
	server := NewServer()
	server.ListenQuietly("0.0.0.0:8080")

	server.Close()

	client := NewClient("0.0.0.0:8080")

	select {
	case <-client.Incoming:
		t.Error("Expected server to be closed")
	case <-time.After(time.Millisecond):
	}
}
