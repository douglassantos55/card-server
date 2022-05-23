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

func TestOthersRemainInQueue(t *testing.T) {
	manager := NewQueueManager()
	dispatcher := NewTestDispatcher()

	p1 := NewTestPlayer()
	p2 := NewTestPlayer()
	p3 := NewTestPlayer()

	go manager.Process(Event{
		Type:   QueueUp,
		Player: p1,
	}, dispatcher)

	go manager.Process(Event{
		Type:   QueueUp,
		Player: p2,
	}, dispatcher)

	go manager.Process(Event{
		Type:   QueueUp,
		Player: p3,
	}, dispatcher)

	<-p1.Outgoing
	<-p2.Outgoing

	select {
	case event := <-dispatcher.Dispatch:
		if event.Type != CreateMatch {
			t.Errorf("Expected %v, got %v", CreateMatch, event.Type)
		}
		players := event.Payload.([]*Player)
		if len(players) != 2 {
			t.Errorf("Expected 2 players, got %v", len(players))
		}
	}

	<-p3.Outgoing

	select {
	case <-time.After(time.Second):
	case <-p3.Outgoing:
		t.Error("Should not receive match found")
	}
}

func TestDequeue(t *testing.T) {
	manager := NewQueueManager()
	player := NewTestPlayer()

	go manager.Process(Event{
		Type:   QueueUp,
		Player: player,
	}, nil)

	<-player.Outgoing

	go manager.Process(Event{
		Type:   Dequeue,
		Player: player,
	}, nil)

	select {
	case <-time.After(time.Second):
		t.Error("Did not receive response from server")
	case response := <-player.Outgoing:
		if response.Type != Dequeued {
			t.Errorf("Expected %v, got %v", Dequeued, response.Type)
		}

		length := len(manager.queue.players)
		if length != 0 {
			t.Errorf("Expected empty queue, got %v", length)
		}
	}
}

func TestDispatchesCreateMatchEvent(t *testing.T) {
	manager := NewQueueManager()
	dispatcher := NewTestDispatcher()

	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	go manager.Process(Event{
		Type:   QueueUp,
		Player: p1,
	}, dispatcher)

	<-p1.Outgoing // wait for match

	go manager.Process(Event{
		Type:   QueueUp,
		Player: p2,
	}, dispatcher)

	<-p2.Outgoing // wait for match

	select {
	case event := <-dispatcher.Dispatch:
		if event.Type != CreateMatch {
			t.Errorf("Expected %v, got %v", CreateMatch, event.Type)
		}
		players := event.Payload.([]*Player)
		if len(players) != 2 {
			t.Errorf("Expected 2 players, got %v", len(players))
		}
	case <-time.After(time.Second):
		t.Error("Expected response from server")
	}
}
