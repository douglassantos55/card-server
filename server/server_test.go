package server

import (
	"testing"
	"time"
)

func TestAcceptsConnections(t *testing.T) {
	server := NewServer(NewDispatcher())
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
	server := NewServer(NewDispatcher())
	server.ListenQuietly("0.0.0.0:8080")

	server.Close()

	client := NewClient("0.0.0.0:8080")

	select {
	case <-client.Incoming:
		t.Error("Expected server to be closed")
	case <-time.After(time.Millisecond):
	}
}

func TestDispatchesEvents(t *testing.T) {
	dispatcher := NewDispatcher()

	handler := &TestHandler{
		make(chan bool),
	}

	dispatcher.Register <- handler

	server := NewServer(dispatcher)
	server.ListenQuietly("0.0.0.0:8080")

	defer server.Close()

	client := NewClient("0.0.0.0:8080")

	client.Outgoing <- Event{
		Type: QueueUp,
	}

	select {
	case executed := <-handler.Executed:
		if !executed {
			t.Error("Expected handler to be executed")
		}
	case <-time.After(time.Second):
		t.Error("Expected response from server")
	}
}
