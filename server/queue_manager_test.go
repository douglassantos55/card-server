package server

import (
	"testing"
	"time"
)

func NewTestPlayer() *Player {
	return &Player{
		Incoming: make(chan Event),
		Outgoing: make(chan Response),
	}
}

func TestReceivesWaitForMatch(t *testing.T) {
	manager := NewQueueManager()
	player := NewTestPlayer()

	go manager.Process(Event{
		Type:   QueueUp,
		Player: player,
	}, nil)

	select {
	case msg := <-player.Outgoing:
		expected := WaitForMatch
		got := msg.Type

		if got != expected {
			t.Errorf("Expecetd %v, got %v", expected, got)
		}
	case <-time.After(time.Second):
		t.Error("Expected response from server")
	}
}

func TestMatchFound(t *testing.T) {
	manager := NewQueueManager()
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	go manager.Process(Event{
		Type:   QueueUp,
		Player: p1,
	}, nil)
	go manager.Process(Event{
		Type:   QueueUp,
		Player: p2,
	}, nil)

	<-p1.Outgoing
	<-p2.Outgoing

	select {
	case response := <-p1.Outgoing:
		expected := MatchFound
		if response.Type != expected {
			t.Errorf("Expected %v, got %v", expected, response.Type)
		}
	case <-time.After(time.Second):
		t.Error("Expected response from server")
	}

	select {
	case response := <-p2.Outgoing:
		expected := MatchFound
		if response.Type != expected {
			t.Errorf("Expected %v, got %v", expected, response.Type)
		}
	case <-time.After(time.Second):
		t.Error("Expected response from server")
	}
}

func TestOthersRemainInQueue(t *testing.T) {
	manager := NewQueueManager()
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()
	p3 := NewTestPlayer()

	go manager.Process(Event{
		Type:   QueueUp,
		Player: p1,
	}, nil)
	go manager.Process(Event{
		Type:   QueueUp,
		Player: p2,
	}, nil)
	go manager.Process(Event{
		Type:   QueueUp,
		Player: p3,
	}, nil)

	<-p1.Outgoing
	<-p2.Outgoing
	<-p3.Outgoing

	select {
	case response := <-p1.Outgoing:
		expected := MatchFound
		if response.Type != expected {
			t.Errorf("Expected %v, got %v", expected, response.Type)
		}
	case <-time.After(time.Second):
		t.Error("Expected response from server")
	}

	select {
	case response := <-p2.Outgoing:
		expected := MatchFound
		if response.Type != expected {
			t.Errorf("Expected %v, got %v", expected, response.Type)
		}
	case <-time.After(time.Second):
		t.Error("Expected response from server")
	}

	select {
	case <-time.After(time.Second):
	case <-p3.Outgoing:
		t.Error("Should not receive match found")
	}
}
