package server

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSendsMatchFoundToPlayers(t *testing.T) {
	maker := NewMatchmaker()
	dispatcher := NewDispatcher()

	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	go maker.Process(Event{
		Type:    CreateMatch,
		Payload: []*Player{p1, p2},
	}, dispatcher)

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

func TestRegistersMatchAsHandler(t *testing.T) {
	maker := NewMatchmaker()
	dispatcher := NewTestDispatcher()

	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	go maker.Process(Event{
		Type:    CreateMatch,
		Payload: []*Player{p1, p2},
	}, dispatcher)

	match := <-dispatcher.Register

	if match == nil {
		t.Error("Expected match to be registered as handler")
	}
}

func TestConfirmMatch(t *testing.T) {
	maker := NewMatchmaker()
	dispatcher := NewDispatcher()

	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	go maker.Process(Event{
		Type:    CreateMatch,
		Payload: []*Player{p1, p2},
	}, dispatcher)

	res := <-p1.Outgoing
	<-p2.Outgoing

	uuid := res.Payload.(uuid.UUID)

	dispatcher.Dispatch <- Event{
		Type:    MatchConfirmed,
		Player:  p2,
		Payload: uuid.String(),
	}

	select {
	case response := <-p2.Outgoing:
		if response.Type != WaitOtherPlayers {
			t.Errorf("Expected %v, got %v", WaitOtherPlayers, response.Type)
		}
	case <-time.After(time.Second):
		t.Error("Expected server response")
	}
}

func TestRefuseMatch(t *testing.T) {
	maker := NewMatchmaker()
	dispatcher := NewDispatcher()

	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	go maker.Process(Event{
		Type:    CreateMatch,
		Payload: []*Player{p1, p2},
	}, dispatcher)

	res := <-p1.Outgoing
	<-p2.Outgoing

	uuid := res.Payload.(uuid.UUID)

	dispatcher.Dispatch <- Event{
		Type:    MatchDeclined,
		Player:  p2,
		Payload: uuid.String(),
	}

	select {
	case response := <-p1.Outgoing:
		if response.Type != MatchCanceled {
			t.Errorf("Expected %v, got %v", MatchCanceled, response.Type)
		}
	case <-time.After(time.Second):
		t.Error("Expected response from server")
	}
}

func TestConfirmedPlayersGoBackToQueue(t *testing.T) {
	maker := NewMatchmaker()

	dispatcher := NewDispatcher()
	dispatcher.Register <- NewQueueManager()

	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	go maker.Process(Event{
		Type:    CreateMatch,
		Payload: []*Player{p1, p2},
	}, dispatcher)

	res := <-p1.Outgoing // match found
	<-p2.Outgoing        // match found

	uuid := res.Payload.(uuid.UUID)

	dispatcher.Dispatch <- Event{
		Type:    MatchConfirmed,
		Player:  p2,
		Payload: uuid.String(),
	}

	<-p2.Outgoing // wait other players

	dispatcher.Dispatch <- Event{
		Type:    MatchDeclined,
		Player:  p1,
		Payload: uuid.String(),
	}

	<-p1.Outgoing // match canceled
	<-p2.Outgoing // match canceled

	select {
	case requeue := <-p2.Outgoing:
		if requeue.Type != WaitForMatch {
			t.Errorf("Expected %v, got %v", QueueUp, requeue.Type)
		}
	case <-time.After(time.Second):
		t.Error("Expected queue up event to be dispatched")
	}

	select {
	case <-p1.Outgoing:
		t.Error("Should not receive response from server")
	case <-time.After(time.Second):
	}
}

func TestMatchIsCanceledIfPlayersDontConfirm(t *testing.T) {
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	dispatcher := NewDispatcher()
	match := NewMatch([]*Player{p1, p2}, 5*time.Millisecond)

	dispatcher.Register <- match

	dispatcher.Dispatch <- Event{
		Type:    AskConfirmation,
		Payload: match.Id,
	}

	<-p1.Outgoing
	<-p2.Outgoing

	select {
	case res := <-p1.Outgoing:
		if res.Type != MatchCanceled {
			t.Errorf("Expected %v, got %v", MatchCanceled, res.Type)
		}
	case <-time.After(6 * time.Millisecond):
		t.Error("Expected response from server")
	}

	select {
	case res := <-p2.Outgoing:
		if res.Type != MatchCanceled {
			t.Errorf("Expected %v, got %v", MatchCanceled, res.Type)
		}
	case <-time.After(6 * time.Millisecond):
		t.Error("Expected response from server")
	}
}

func TestDispatchesStartGame(t *testing.T) {
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	dispatcher := NewTestDispatcher()
	match := NewMatch([]*Player{p1, p2}, time.Minute)

	go match.Process(Event{
		Type:    MatchConfirmed,
		Player:  p1,
		Payload: match.Id.String(),
	}, dispatcher)

	<-p1.Outgoing // wait other players

	go match.Process(Event{
		Type:    MatchConfirmed,
		Player:  p2,
		Payload: match.Id.String(),
	}, dispatcher)

	<-p2.Outgoing // wait other players

	select {
	case event := <-dispatcher.Dispatch:
		if event.Type != StartGame {
			t.Errorf("Expected %v, got %v", StartGame, event.Type)
		}
	case <-time.After(time.Second):
		t.Error("Expected event to be dispatched")
	}
}

func TestMatchIsRemovedFromDispatcher(t *testing.T) {

}
