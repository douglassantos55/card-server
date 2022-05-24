package server

import (
	"testing"
	"time"
)

func TestRegistersGameAsHandler(t *testing.T) {
	manager := NewGameManager()
	dispatcher := NewTestDispatcher()

	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	go manager.Process(Event{
		Type:    StartGame,
		Payload: []*Player{p1, p2},
	}, dispatcher)

	select {
	case <-dispatcher.Register:
	case <-time.After(time.Second):
		t.Error("Expected game to be registered as handler")
	}
}
