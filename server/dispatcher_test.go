package server

import (
	"testing"
)

type TestHandler struct {
	Executed chan bool
}

func (t *TestHandler) Process(event Event, dispatcher *Dispatcher) {
	t.Executed <- true
}

func TestRegisterHandlers(t *testing.T) {
	dispatcher := NewDispatcher()
	handler := &TestHandler{
		Executed: make(chan bool),
	}

	dispatcher.Register <- handler

	dispatcher.Dispatch <- Event{
		Type: QueueUp,
	}

	if !<-handler.Executed {
		t.Error("Expected handler to be executed")
	}
}
