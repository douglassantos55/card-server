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

func TestSendsThreeCardsAsStartingHand(t *testing.T) {
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	game := NewGame([]*Player{p1, p2})
	go game.Start(time.Minute)

	select {
	case <-time.After(time.Second):
		t.Error("Expected starting hand")
	case response := <-p1.Outgoing:
		if response.Type != StartingHand {
			t.Errorf("Expected %v, got %v", StartingHand, response.Type)
		}
		cards := response.Payload.([]HasManaCost)
		if len(cards) != 3 {
			t.Errorf("Expected %v cards, got %v", 3, len(cards))
		}
	}

	select {
	case <-time.After(time.Second):
		t.Error("Expected starting hand")
	case response := <-p2.Outgoing:
		if response.Type != StartingHand {
			t.Errorf("Expected %v, got %v", StartingHand, response.Type)
		}
		cards := response.Payload.([]HasManaCost)
		if len(cards) != 3 {
			t.Errorf("Expected %v cards, got %v", 3, len(cards))
		}
	}
}

func TestReplacesDiscardedCards(t *testing.T) {
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	game := NewGame([]*Player{p1, p2})

	dispatcher := NewDispatcher()
	dispatcher.Register <- game

	go game.Start(time.Minute)

	res1 := <-p1.Outgoing
	hand1 := res1.Payload.([]HasManaCost)

	<-p2.Outgoing

	dispatcher.Dispatch <- Event{
		Type:   CardsDiscarded,
		Player: p1,
		Payload: CardsDiscardedPayload{
			GameId: game.Id.String(),
			Cards: []string{
				hand1[1].GetId(),
				hand1[0].GetId(),
			},
		},
	}

	select {
	case <-time.After(time.Second):
		t.Error("Expected replacements")
	case response := <-p1.Outgoing:
		if response.Type != WaitOtherPlayers {
			t.Errorf("Expected %v, got %v", WaitOtherPlayers, response.Type)
		}

		cards := response.Payload.([]HasManaCost)
		if len(cards) != 3 {
			t.Errorf("Expected %v, got %v", 3, len(cards))
		}
	}
}

func TestTimerToChooseStartingHand(t *testing.T) {
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	game := NewGame([]*Player{p1, p2})

	dispatcher := NewDispatcher()
	dispatcher.Register <- game

	go game.Start(100 * time.Millisecond)

	<-p1.Outgoing // starting hand
	<-p2.Outgoing // starting hand

	time.Sleep(200 * time.Millisecond)

	select {
	case response := <-p1.Outgoing:
		if response.Type != StartTurn {
			t.Errorf("Expected %v, got %v", StartTurn, response.Type)
		}
	case <-time.After(time.Second):
		t.Error("Expected start turn")
	}

	select {
	case response := <-p2.Outgoing:
		if response.Type != WaitTurn {
			t.Errorf("Expected %v, got %v", WaitTurn, response.Type)
		}
	case <-time.After(time.Second):
		t.Error("Expected wait turn")
	}
}

func TestConfirmWithoutDiscarding(t *testing.T) {
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	game := NewGame([]*Player{p1, p2})

	dispatcher := NewDispatcher()
	dispatcher.Register <- game

	go game.Start(time.Minute)

	res := <-p1.Outgoing
	<-p2.Outgoing

	hand := res.Payload.([]HasManaCost)

	dispatcher.Dispatch <- Event{
		Type:   CardsDiscarded,
		Player: p1,
		Payload: CardsDiscardedPayload{
			GameId: game.Id.String(),
			Cards:  []string{},
		},
	}

	select {
	case <-time.After(time.Second):
		t.Error("Expected response from server")
	case response := <-p1.Outgoing:
		if response.Type != WaitOtherPlayers {
			t.Errorf("Expected %v, got %v", WaitOtherPlayers, response.Type)
		}

		cards := response.Payload.([]HasManaCost)
		if len(cards) != 3 {
			t.Errorf("Expected %v, got %v", 3, len(cards))
		}

		count := 0
		for _, card := range hand {
			for _, c := range cards {
				if card.GetId() == c.GetId() {
					count++
				}
			}
		}
		if count != 3 {
			t.Error("Expected same cards")
		}
	}
}

func TestStartsTurnsWhenBothPlayersChoseHand(t *testing.T) {
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	game := NewGame([]*Player{p1, p2})

	dispatcher := NewDispatcher()
	dispatcher.Register <- game

	go game.Start(time.Minute)

	res1 := <-p1.Outgoing
	hand1 := res1.Payload.([]HasManaCost)

	res2 := <-p2.Outgoing
	hand2 := res2.Payload.([]HasManaCost)

	dispatcher.Dispatch <- Event{
		Type:   CardsDiscarded,
		Player: p2,
		Payload: CardsDiscardedPayload{
			GameId: game.Id.String(),
			Cards: []string{
				hand2[1].GetId(),
				hand2[0].GetId(),
			},
		},
	}

	select {
	case <-time.After(time.Second):
		t.Error("Expected replacements")
	case response := <-p2.Outgoing:
		if response.Type != WaitOtherPlayers {
			t.Errorf("Expected %v, got %v", WaitOtherPlayers, response.Type)
		}
		cards := response.Payload.([]HasManaCost)
		if len(cards) != 3 {
			t.Errorf("Expected %v, got %v", 3, len(cards))
		}
	}

	dispatcher.Dispatch <- Event{
		Type:   CardsDiscarded,
		Player: p1,
		Payload: CardsDiscardedPayload{
			GameId: game.Id.String(),
			Cards: []string{
				hand1[2].GetId(),
			},
		},
	}

	<-p1.Outgoing // wait other players

	select {
	case <-time.After(time.Second):
		t.Error("Expected turn start")
	case response := <-p1.Outgoing:
		if response.Type != StartTurn {
			t.Errorf("Expected %v, got %v", StartTurn, response.Type)
		}
	}

	select {
	case <-time.After(time.Second):
		t.Error("Expected wait turn")
	case response := <-p2.Outgoing:
		if response.Type != WaitTurn {
			t.Errorf("Expected %v, got %v", WaitTurn, response.Type)
		}
	}
}
