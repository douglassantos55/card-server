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

		payload := response.Payload.(StartingHandPayload)
		if payload.GameId != game.Id {
			t.Errorf("Expected %v, got %v", game.Id, payload.GameId)
		}
		if len(payload.Cards) != 3 {
			t.Errorf("Expected %v cards, got %v", 3, len(payload.Cards))
		}
	}

	select {
	case <-time.After(time.Second):
		t.Error("Expected starting hand")
	case response := <-p2.Outgoing:
		if response.Type != StartingHand {
			t.Errorf("Expected %v, got %v", StartingHand, response.Type)
		}

		payload := response.Payload.(StartingHandPayload)
		if payload.GameId != game.Id {
			t.Errorf("Expected %v, got %v", game.Id, payload.GameId)
		}
		if len(payload.Cards) != 3 {
			t.Errorf("Expected %v cards, got %v", 3, len(payload.Cards))
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
	hand1 := res1.Payload.(StartingHandPayload)

	<-p2.Outgoing

	dispatcher.Dispatch <- Event{
		Type:   CardsDiscarded,
		Player: p1,
		Payload: CardsDiscardedPayload{
			GameId: game.Id.String(),
			Cards: []string{
				hand1.Cards[1].GetId(),
				hand1.Cards[0].GetId(),
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

	hand := res.Payload.(StartingHandPayload)

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
		for _, card := range hand.Cards {
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
	hand1 := res1.Payload.(StartingHandPayload)

	res2 := <-p2.Outgoing
	hand2 := res2.Payload.(StartingHandPayload)

	dispatcher.Dispatch <- Event{
		Type:   CardsDiscarded,
		Player: p2,
		Payload: CardsDiscardedPayload{
			GameId: game.Id.String(),
			Cards: []string{
				hand2.Cards[1].GetId(),
				hand2.Cards[0].GetId(),
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
				hand1.Cards[2].GetId(),
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

		payload := response.Payload.(TurnPayload)

		if payload.Duration != 75*time.Second {
			t.Errorf("Expected %v, got %v", 75*time.Second, payload.Duration)
		}

		if payload.GameId != game.Id {
			t.Errorf("Expected %v, got %v", game.Id, payload.GameId)
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

func TestTurnTimer(t *testing.T) {
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	game := NewGame([]*Player{p1, p2})
	go game.StartTurns(100 * time.Millisecond)

	<-p1.Outgoing // start turn
	<-p2.Outgoing // wait turn

	time.Sleep(150 * time.Millisecond)

	select {
	case response := <-p1.Outgoing:
		if response.Type != WaitTurn {
			t.Errorf("Expected %v, got %v", WaitTurn, response.Type)
		}
	case <-time.After(time.Second):
		t.Error("Expected turn to end")
	}

	select {
	case response := <-p2.Outgoing:
		if response.Type != StartTurn {
			t.Errorf("Expected %v, got %v", StartTurn, response.Type)
		}
	case <-time.After(time.Second):
		t.Error("Expected turn to start")
	}
}

func TestEndTurn(t *testing.T) {
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	dispatcher := NewDispatcher()
	game := NewGame([]*Player{p1, p2})

	dispatcher.Register <- game
	go game.StartTurns(time.Minute)

	res := <-p1.Outgoing
	<-p2.Outgoing

	payload := res.Payload.(TurnPayload)

	dispatcher.Dispatch <- Event{
		Type:    EndTurn,
		Player:  p1,
		Payload: payload.GameId.String(),
	}

	select {
	case response := <-p1.Outgoing:
		if response.Type != WaitTurn {
			t.Errorf("Expected %v, got %v", WaitTurn, response.Type)
		}
	case <-time.After(time.Second):
		t.Error("Expected turn to end")
	}

	select {
	case response := <-p2.Outgoing:
		if response.Type != StartTurn {
			t.Errorf("Expected %v, got %v", StartTurn, response.Type)
		}
		payload := response.Payload.(TurnPayload)
		if payload.Duration != time.Minute {
			t.Errorf("Expected %v, got %v", time.Minute, payload.Duration)
		}
		if payload.GameId != game.Id {
			t.Errorf("Expecetd %v, got %v", game.Id, payload.GameId)
		}
	case <-time.After(time.Second):
		t.Error("Expected turn to start")
	}

	dispatcher.Dispatch <- Event{
		Type:    EndTurn,
		Player:  p2,
		Payload: payload.GameId.String(),
	}

	select {
	case response := <-p1.Outgoing:
		if response.Type != StartTurn {
			t.Errorf("Expected %v, got %v", StartTurn, response.Type)
		}
	case <-time.After(time.Second):
		t.Error("Expected turn to start")
	}
}

func TestDrawsCardOnTurnStart(t *testing.T) {
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	game := NewGame([]*Player{p1, p2})
	go game.Start(100 * time.Millisecond)

	<-p1.Outgoing // starting hand
	<-p2.Outgoing // starting hand

	time.Sleep(200 * time.Millisecond)

	select {
	case response := <-p1.Outgoing:
		payload := response.Payload.(TurnPayload)

		// 3 from starting hand + 1 from star turn
		if payload.CardsLeft != 56 {
			t.Errorf("Expected %v, got %v", 56, payload.CardsLeft)
		}

		if payload.Card == nil {
			t.Error("Expected card to be drawn")
		}
	}
}

func TestPlayCard(t *testing.T) {
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	dispatcher := NewDispatcher()
	game := NewGame([]*Player{p1, p2})

	dispatcher.Register <- game
	go game.StartTurns(time.Minute)

	res := <-p1.Outgoing
	<-p2.Outgoing

	payload := res.Payload.(TurnPayload)

	// reduce all of it so we can actually play any card
	payload.Card.ReduceManaCost(100)

	dispatcher.Dispatch <- Event{
		Type:   PlayCard,
		Player: p1,
		Payload: PlayCardPayload{
			Card:   payload.Card.GetId(),
			GameId: game.Id.String(),
		},
	}

	select {
	case <-time.After(500 * time.Millisecond):
		t.Error("Expected card played response")
	case response := <-p1.Outgoing:
		if response.Type != CardPlayed {
			t.Errorf("Expected %v, got %v", CardPlayed, response.Type)
		}

		got := response.Payload.(CardPlayedPayload)

		if got.Card.GetId() != payload.Card.GetId() {
			t.Errorf("Expected %v, got %v", payload.Card, got)
		}

		if got.Card.GetStatus().CanAttack() {
			t.Error("Should be exhausted")
		}
	}

	select {
	case <-time.After(500 * time.Millisecond):
		t.Error("Expected card played response")
	case response := <-p2.Outgoing:
		if response.Type != CardPlayed {
			t.Errorf("Expected %v, got %v", CardPlayed, response.Type)
		}

		got := response.Payload.(CardPlayedPayload)

		if got.Card.GetId() != payload.Card.GetId() {
			t.Errorf("Expected %v, got %v", payload.Card, got)
		}

		if got.Card.GetStatus().CanAttack() {
			t.Error("Should be exhausted")
		}
	}
}

func TestGainsManaOnTurnStart(t *testing.T) {
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	game := NewGame([]*Player{p1, p2})
	go game.StartTurns(100 * time.Millisecond)

	res := <-p1.Outgoing
	<-p2.Outgoing // wait turn

	payload := res.Payload.(TurnPayload)

	if payload.Mana != 1 {
		t.Errorf("Expected %v, got %v", 1, payload.Mana)
	}

	// wait turn end
	time.Sleep(100 * time.Millisecond)

	<-p1.Outgoing // wait turn
	res2 := <-p2.Outgoing

	payload2 := res2.Payload.(TurnPayload)
	if payload2.Mana != 1 {
		t.Errorf("Expected %v, got %v", 1, payload2.Mana)
	}

	// wait turn end
	time.Sleep(100 * time.Millisecond)

	res3 := <-p1.Outgoing
	<-p2.Outgoing // wait turn

	payload3 := res3.Payload.(TurnPayload)

	if payload3.Mana != 2 {
		t.Errorf("Expected %v, got %v", 2, payload3.Mana)
	}
}

func TestCannotPlayCardWithManaCostHigherThanCurrentMana(t *testing.T) {
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	dispatcher := NewDispatcher()
	game := NewGame([]*Player{p1, p2})

	dispatcher.Register <- game
	go game.StartTurns(time.Minute)

	res := <-p1.Outgoing // start turn
	<-p2.Outgoing        // wait turn

	payload := res.Payload.(TurnPayload)

	payload.Card.IncreaseManaCost(100)

	dispatcher.Dispatch <- Event{
		Type:   PlayCard,
		Player: p1,
		Payload: PlayCardPayload{
			GameId: game.Id.String(),
			Card:   payload.Card.GetId(),
		},
	}

	select {
	case <-time.After(time.Second):
		t.Error("Expected error response")
	case res := <-p1.Outgoing:
		if res.Type != Error {
			t.Errorf("Expected %v, got %v", Error, res.Type)
		}

		expected := "Not enough mana"
		received := res.Payload.(string)

		if received != expected {
			t.Errorf("Expected '%v', got '%v'", expected, received)
		}
	}
}

func TestPlayedCardIsRemovedFromHand(t *testing.T) {
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	game := NewGame([]*Player{p1, p2})
	go game.StartTurns(time.Minute)

	res := <-p1.Outgoing // start turn
	<-p2.Outgoing        // wait turn

	payload := res.Payload.(TurnPayload)
	payload.Card.ReduceManaCost(100)

	go game.Process(Event{
		Type:   PlayCard,
		Player: p1,
		Payload: PlayCardPayload{
			GameId: game.Id.String(),
			Card:   payload.Card.GetId(),
		},
	}, nil)

	<-p1.Outgoing // card played
	<-p2.Outgoing // card played

	go game.Process(Event{
		Type:    EndTurn,
		Player:  p1,
		Payload: game.Id.String(),
	}, nil)

	<-p1.Outgoing // wait turn
	<-p2.Outgoing // start turn

	go game.Process(Event{
		Type:    EndTurn,
		Player:  p2,
		Payload: game.Id.String(),
	}, nil)

	res2 := <-p1.Outgoing // start turn
	<-p2.Outgoing         // wait turn

	payload2 := res2.Payload.(TurnPayload)

	if payload2.CardsInHand != 4 {
		t.Errorf("Expected %v, got %v", 4, payload2.CardsInHand)
	}
}

func TestPlayingCardsUsesMana(t *testing.T) {
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	game := NewGame([]*Player{p1, p2})
	go game.Start(100 * time.Millisecond)

	res := <-p1.Outgoing // starting hand
	<-p2.Outgoing        // starting hand

	hand := res.Payload.(StartingHandPayload)

	// start turns
	time.Sleep(110 * time.Millisecond)

	turn := <-p1.Outgoing // start turn
	<-p2.Outgoing         // wait turn

	payload := turn.Payload.(TurnPayload)
	payload.Card.ReduceManaCost(payload.Card.GetManaCost() - 1)

	go game.Process(Event{
		Type:   PlayCard,
		Player: p1,
		Payload: PlayCardPayload{
			GameId: game.Id.String(),
			Card:   payload.Card.GetId(),
		},
	}, nil)

	<-p1.Outgoing // card played
	<-p2.Outgoing // card played

	hand.Cards[0].ReduceManaCost(hand.Cards[0].GetManaCost() - 1)

	go game.Process(Event{
		Type:   PlayCard,
		Player: p1,
		Payload: PlayCardPayload{
			GameId: game.Id.String(),
			Card:   hand.Cards[0].GetId(),
		},
	}, nil)

	select {
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected error response")
	case res := <-p1.Outgoing:
		if res.Type != Error {
			t.Errorf("Expected %v, got %v", Error, res.Type)
		}
		expected := "Not enough mana"
		received := res.Payload.(string)

		if received != expected {
			t.Errorf("Expected %v, got %v", expected, received)
		}
	}

	select {
	case <-time.After(100 * time.Millisecond):
	case res := <-p2.Outgoing:
		t.Errorf("Should not receive, got %v", res)
	}
}

func TestRefillsManaOnTurnStart(t *testing.T) {
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	game := NewGame([]*Player{p1, p2})
	go game.StartTurns(100 * time.Millisecond)

	<-p1.Outgoing // start turn
	<-p2.Outgoing // wait turn

	time.Sleep(100 * time.Millisecond)

	<-p1.Outgoing // wait turn
	<-p2.Outgoing // start turn

	time.Sleep(100 * time.Millisecond)

	res := <-p1.Outgoing // start turn
	<-p2.Outgoing        // wait turn

	payload := res.Payload.(TurnPayload)
	payload.Card.ReduceManaCost(payload.Card.GetManaCost() - 2)

	go game.Process(Event{
		Type:   PlayCard,
		Player: p1,
		Payload: PlayCardPayload{
			GameId: game.Id.String(),
			Card:   payload.Card.GetId(),
		},
	}, nil)

	<-p1.Outgoing // card played
	<-p2.Outgoing // card played

	time.Sleep(100 * time.Millisecond)

	<-p1.Outgoing // wait turn
	<-p2.Outgoing // start turn

	time.Sleep(100 * time.Millisecond)

	res2 := <-p1.Outgoing // start turn
	<-p2.Outgoing         // wait turn

	payload2 := res2.Payload.(TurnPayload)

	if payload2.Mana != 3 {
		t.Errorf("Expected %v, got %v", 3, payload2.Mana)
	}
}

func TestCapManaTo10(t *testing.T) {
	player := &GamePlayer{}

	player.IncreaseMana(8)
	if player.MaxMana != 8 {
		t.Errorf("Expected %v, got %v", 8, player.MaxMana)
	}

	player.IncreaseMana(100)
	if player.MaxMana != 10 {
		t.Errorf("Expected %v, got %v", 10, player.MaxMana)
	}

	player.GainMana(300)
	if player.Mana != 10 {
		t.Errorf("Expected %v, got %v", 10, player.Mana)
	}

	player.ConsumeMana(1000)
	if player.Mana != 0 {
		t.Errorf("Expected %v, got %v", 0, player.Mana)
	}

	player.RefillMana()
	if player.Mana != 10 {
		t.Errorf("Expected %v, got %v", 10, player.Mana)
	}
}

func TestAttackCards(t *testing.T) {
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	game := NewGame([]*Player{p1, p2})
	go game.StartTurns(time.Minute)

	res := <-p1.Outgoing // start turn
	<-p2.Outgoing        // wait turn

	payload := res.Payload.(TurnPayload)

	attacker := payload.Card.(Minion)

	attacker.ReduceManaCost(100)
	attacker.ReduceHealth(100)
	attacker.ReduceDamage(100)

	attacker.GainDamage(2)
	attacker.GainHealth(2)

	go game.Process(Event{
		Type:   PlayCard,
		Player: p1,
		Payload: PlayCardPayload{
			GameId: game.Id.String(),
			Card:   payload.Card.GetId(),
		},
	}, nil)

	<-p1.Outgoing // card played
	<-p2.Outgoing // card played

	go game.Process(Event{
		Type:    EndTurn,
		Player:  p1,
		Payload: game.Id.String(),
	}, nil)

	<-p1.Outgoing         // wait turn
	res2 := <-p2.Outgoing // start turn

	payload2 := res2.Payload.(TurnPayload)

	defender := payload2.Card.(Minion)
	defender.ReduceManaCost(100)
	defender.ReduceDamage(defender.GetDamage() - 1)
	defender.ReduceHealth(defender.GetHealth() - 3)

	go game.Process(Event{
		Type:   PlayCard,
		Player: p2,
		Payload: PlayCardPayload{
			GameId: game.Id.String(),
			Card:   payload2.Card.GetId(),
		},
	}, nil)

	<-p1.Outgoing // card played
	<-p2.Outgoing // card played

	go game.Process(Event{
		Type:    EndTurn,
		Player:  p2,
		Payload: game.Id.String(),
	}, nil)

	<-p1.Outgoing // start turn
	<-p2.Outgoing // wait turn

	go game.Process(Event{
		Type:   Attack,
		Player: p1,
		Payload: AttackPayload{
			GameId:   game.Id.String(),
			Attacker: payload.Card.GetId(),
			Target:   payload2.Card.GetId(),
		},
	}, nil)

	select {
	case <-time.After(500 * time.Millisecond):
		t.Error("Expected attack result")
	case res := <-p1.Outgoing:
		if res.Type != AttackResult {
			t.Errorf("Expected %v, got %v", AttackResult, res.Type)
		}

		boards := res.Payload.([]*Board)

		attacker := boards[0].Defenders[attacker.GetId()]
		if attacker.GetHealth() != 1 {
			t.Errorf("Expected %v, got %v", 1, attacker.GetHealth())
		}

		defender := boards[1].Defenders[defender.GetId()]
		if defender.GetHealth() != 1 {
			t.Errorf("Expected %v, got %v", 1, defender.GetHealth())
		}
	}

	select {
	case <-time.After(500 * time.Millisecond):
		t.Error("Expected attack result")
	case res := <-p2.Outgoing:
		if res.Type != AttackResult {
			t.Errorf("Expected %v, got %v", AttackResult, res.Type)
		}

		boards := res.Payload.([]*Board)
		attacker := boards[1].Defenders[attacker.GetId()]
		if attacker.GetHealth() != 1 {
			t.Errorf("Expected %v, got %v", 1, attacker.GetHealth())
		}

		defender := boards[0].Defenders[defender.GetId()]
		if defender.GetHealth() != 1 {
			t.Errorf("Expected %v, got %v", 1, defender.GetHealth())
		}
	}
}

func TestDestroysDefender(t *testing.T) {
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	game := NewGame([]*Player{p1, p2})
	go game.StartTurns(time.Minute)

	res := <-p1.Outgoing // start turn
	<-p2.Outgoing        // wait turn

	payload := res.Payload.(TurnPayload)

	attacker := payload.Card.(Minion)

	attacker.ReduceManaCost(100)
	attacker.ReduceHealth(100)
	attacker.ReduceDamage(100)

	attacker.GainDamage(2)
	attacker.GainHealth(2)

	go game.Process(Event{
		Type:   PlayCard,
		Player: p1,
		Payload: PlayCardPayload{
			GameId: game.Id.String(),
			Card:   payload.Card.GetId(),
		},
	}, nil)

	<-p1.Outgoing // card played
	<-p2.Outgoing // card played

	go game.Process(Event{
		Type:    EndTurn,
		Player:  p1,
		Payload: game.Id.String(),
	}, nil)

	<-p1.Outgoing         // wait turn
	res2 := <-p2.Outgoing // start turn

	payload2 := res2.Payload.(TurnPayload)

	defender := payload2.Card.(Minion)
	defender.ReduceManaCost(100)
	defender.ReduceDamage(defender.GetDamage() - 1)
	defender.ReduceHealth(defender.GetHealth() - 1)

	go game.Process(Event{
		Type:   PlayCard,
		Player: p2,
		Payload: PlayCardPayload{
			GameId: game.Id.String(),
			Card:   payload2.Card.GetId(),
		},
	}, nil)

	<-p1.Outgoing // card played
	<-p2.Outgoing // card played

	go game.Process(Event{
		Type:    EndTurn,
		Player:  p2,
		Payload: game.Id.String(),
	}, nil)

	<-p1.Outgoing // start turn
	<-p2.Outgoing // wait turn

	go game.Process(Event{
		Type:   Attack,
		Player: p1,
		Payload: AttackPayload{
			GameId:   game.Id.String(),
			Attacker: payload.Card.GetId(),
			Target:   payload2.Card.GetId(),
		},
	}, nil)

	select {
	case <-time.After(500 * time.Millisecond):
		t.Error("Expected attack result")
	case res := <-p1.Outgoing:
		if res.Type != AttackResult {
			t.Errorf("Expected %v, got %v", AttackResult, res.Type)
		}

		boards := res.Payload.([]*Board)

		attacker := boards[0].Defenders[attacker.GetId()]
		if attacker.GetHealth() != 2 {
			t.Errorf("Expected %v, got %v", 2, attacker.GetHealth())
		}

		if len(boards[1].Defenders) != 0 {
			t.Errorf("Expected %v cards, got %v", 0, len(boards[1].Defenders))
		}
	}

	select {
	case <-time.After(500 * time.Millisecond):
		t.Error("Expected attack result")
	case res := <-p2.Outgoing:
		if res.Type != AttackResult {
			t.Errorf("Expected %v, got %v", AttackResult, res.Type)
		}

		boards := res.Payload.([]*Board)

		if len(boards[0].Defenders) != 0 {
			t.Errorf("Expected %v cards, got %v", 0, len(boards[1].Defenders))
		}

		attacker := boards[1].Defenders[attacker.GetId()]
		if attacker.GetHealth() != 2 {
			t.Errorf("Expected %v, got %v", 2, attacker.GetHealth())
		}
	}
}

func TestDestroysAttacker(t *testing.T) {
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	game := NewGame([]*Player{p1, p2})
	go game.StartTurns(time.Minute)

	res := <-p1.Outgoing // start turn
	<-p2.Outgoing        // wait turn

	payload := res.Payload.(TurnPayload)

	attacker := payload.Card.(Minion)

	attacker.ReduceManaCost(100)
	attacker.ReduceHealth(100)
	attacker.ReduceDamage(100)

	attacker.GainDamage(1)
	attacker.GainHealth(1)

	go game.Process(Event{
		Type:   PlayCard,
		Player: p1,
		Payload: PlayCardPayload{
			GameId: game.Id.String(),
			Card:   payload.Card.GetId(),
		},
	}, nil)

	<-p1.Outgoing // card played
	<-p2.Outgoing // card played

	go game.Process(Event{
		Type:    EndTurn,
		Player:  p1,
		Payload: game.Id.String(),
	}, nil)

	<-p1.Outgoing         // wait turn
	res2 := <-p2.Outgoing // start turn

	payload2 := res2.Payload.(TurnPayload)

	defender := payload2.Card.(Minion)
	defender.ReduceManaCost(100)
	defender.ReduceDamage(defender.GetDamage() - 2)
	defender.ReduceHealth(defender.GetHealth() - 2)

	go game.Process(Event{
		Type:   PlayCard,
		Player: p2,
		Payload: PlayCardPayload{
			GameId: game.Id.String(),
			Card:   payload2.Card.GetId(),
		},
	}, nil)

	<-p1.Outgoing // card played
	<-p2.Outgoing // card played

	go game.Process(Event{
		Type:    EndTurn,
		Player:  p2,
		Payload: game.Id.String(),
	}, nil)

	<-p1.Outgoing // start turn
	<-p2.Outgoing // wait turn

	go game.Process(Event{
		Type:   Attack,
		Player: p1,
		Payload: AttackPayload{
			GameId:   game.Id.String(),
			Attacker: payload.Card.GetId(),
			Target:   payload2.Card.GetId(),
		},
	}, nil)

	select {
	case <-time.After(500 * time.Millisecond):
		t.Error("Expected attack result")
	case res := <-p1.Outgoing:
		if res.Type != AttackResult {
			t.Errorf("Expected %v, got %v", AttackResult, res.Type)
		}

		boards := res.Payload.([]*Board)
		if len(boards[0].Defenders) != 0 {
			t.Errorf("Expected %v cards, got %v", 0, len(boards[1].Defenders))
		}

		defender := boards[1].Defenders[defender.GetId()]
		if defender.GetHealth() != 1 {
			t.Errorf("Expected %v, got %v", 1, defender.GetHealth())
		}
	}

	select {
	case <-time.After(500 * time.Millisecond):
		t.Error("Expected attack result")
	case res := <-p2.Outgoing:
		if res.Type != AttackResult {
			t.Errorf("Expected %v, got %v", AttackResult, res.Type)
		}

		boards := res.Payload.([]*Board)

		defender := boards[0].Defenders[defender.GetId()]
		if defender.GetHealth() != 1 {
			t.Errorf("Expected %v, got %v", 1, defender.GetHealth())
		}

		if len(boards[1].Defenders) != 0 {
			t.Errorf("Expected %v cards, got %v", 0, len(boards[1].Defenders))
		}
	}
}

func TestAttackHero(t *testing.T) {
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	game := NewGame([]*Player{p1, p2})
	go game.StartTurns(time.Minute)

	res := <-p1.Outgoing // start turn
	<-p2.Outgoing        // wait turn

	payload := res.Payload.(TurnPayload)
	payload.Card.ReduceManaCost(payload.Card.GetManaCost() - 1)

	go game.Process(Event{
		Type:   PlayCard,
		Player: p1,
		Payload: PlayCardPayload{
			GameId: game.Id.String(),
			Card:   payload.Card.GetId(),
		},
	}, nil)

	<-p1.Outgoing
	<-p2.Outgoing

	go game.Process(Event{
		Type:    EndTurn,
		Player:  p1,
		Payload: game.Id.String(),
	}, nil)

	<-p1.Outgoing // wait turn
	<-p2.Outgoing // start turn

	go game.Process(Event{
		Type:    EndTurn,
		Player:  p2,
		Payload: game.Id.String(),
	}, nil)

	<-p1.Outgoing // start turn
	<-p2.Outgoing // wait turn

	go game.Process(Event{
		Type:   AttackPlayer,
		Player: p1,
		Payload: AttackPayload{
			GameId:   game.Id.String(),
			Attacker: payload.Card.GetId(),
		},
	}, nil)

	select {
	case <-time.After(500 * time.Millisecond):
		t.Error("expected damage taken response")
	case res := <-p1.Outgoing:
		if res.Type != DamageTaken {
			t.Errorf("EXpected %v, got %v", DamageTaken, res.Type)
		}

		data := res.Payload.(DamageTakenPayload)
		expected := 30 - payload.Card.(Defender).GetDamage()

		if data.Health != expected {
			t.Errorf("Expected %v, got %v", expected, data.Health)
		}
	}

	select {
	case <-time.After(500 * time.Millisecond):
		t.Error("expected damage taken response")
	case res := <-p2.Outgoing:
		if res.Type != DamageTaken {
			t.Errorf("EXpected %v, got %v", DamageTaken, res.Type)
		}

		data := res.Payload.(DamageTakenPayload)
		expected := 30 - payload.Card.(Defender).GetDamage()

		if data.Health != expected {
			t.Errorf("Expected %v, got %v", expected, data.Health)
		}
	}
}

func TestCannotAttackPlayerIfThereAreMinionsOnBoard(t *testing.T) {
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	game := NewGame([]*Player{p1, p2})
	go game.StartTurns(time.Minute)

	res := <-p1.Outgoing // start turn
	<-p2.Outgoing        // wait turn

	payload := res.Payload.(TurnPayload)
	payload.Card.ReduceManaCost(payload.Card.GetManaCost() - 1)

	go game.Process(Event{
		Type:   PlayCard,
		Player: p1,
		Payload: PlayCardPayload{
			GameId: game.Id.String(),
			Card:   payload.Card.GetId(),
		},
	}, nil)

	<-p1.Outgoing // card played
	<-p2.Outgoing // card played

	go game.Process(Event{
		Type:    EndTurn,
		Player:  p1,
		Payload: game.Id.String(),
	}, nil)

	<-p1.Outgoing         // wait turn
	res2 := <-p2.Outgoing // start turn

	payload2 := res2.Payload.(TurnPayload)
	payload2.Card.ReduceManaCost(payload2.Card.GetManaCost() - 1)

	go game.Process(Event{
		Type:   PlayCard,
		Player: p2,
		Payload: PlayCardPayload{
			GameId: game.Id.String(),
			Card:   payload2.Card.GetId(),
		},
	}, nil)

	<-p1.Outgoing // card played
	<-p2.Outgoing // card played

	go game.Process(Event{
		Type:    EndTurn,
		Player:  p2,
		Payload: game.Id.String(),
	}, nil)

	<-p1.Outgoing // start turn
	<-p2.Outgoing // wait turn

	go game.Process(Event{
		Type:   AttackPlayer,
		Player: p1,
		Payload: AttackPayload{
			GameId:   game.Id.String(),
			Attacker: payload.Card.GetId(),
		},
	}, nil)

	select {
	case <-time.After(500 * time.Millisecond):
		t.Error("expected error response")
	case res := <-p1.Outgoing:
		if res.Type != Error {
			t.Errorf("EXpected %v, got %v", Error, res.Type)
		}

		got := res.Payload.(string)
		expected := "Cannot attack player with minions on board"

		if got != expected {
			t.Errorf("Expected %v, got %v", expected, got)
		}
	}
}

func TestGameOver(t *testing.T) {
	p1 := NewTestPlayer()
	p2 := NewTestPlayer()

	game := NewGame([]*Player{p1, p2})
	go game.StartTurns(time.Minute)

	res := <-p1.Outgoing // start turn
	<-p2.Outgoing        // wait turn

	payload := res.Payload.(TurnPayload)
	payload.Card.ReduceManaCost(payload.Card.GetManaCost() - 1)
	payload.Card.(Defender).GainDamage(30)

	go game.Process(Event{
		Type:   PlayCard,
		Player: p1,
		Payload: PlayCardPayload{
			GameId: game.Id.String(),
			Card:   payload.Card.GetId(),
		},
	}, nil)

	<-p1.Outgoing
	<-p2.Outgoing

	go game.Process(Event{
		Type:    EndTurn,
		Player:  p1,
		Payload: game.Id.String(),
	}, nil)

	<-p1.Outgoing // wait turn
	<-p2.Outgoing // start turn

	go game.Process(Event{
		Type:    EndTurn,
		Player:  p2,
		Payload: game.Id.String(),
	}, nil)

	<-p1.Outgoing // start turn
	<-p2.Outgoing // wait turn

	go game.Process(Event{
		Type:   AttackPlayer,
		Player: p1,
		Payload: AttackPayload{
			GameId:   game.Id.String(),
			Attacker: payload.Card.GetId(),
		},
	}, nil)

	select {
	case <-time.After(500 * time.Millisecond):
		t.Error("expected game over response")
	case res := <-p1.Outgoing:
		if res.Type != GameOver {
			t.Errorf("EXpected %v, got %v", GameOver, res.Type)
		}

		payload := res.Payload.(GameOverPayload)
		if payload.Winner.player != p1 {
			t.Error("Wrong winner")
		}
		if payload.Loser.player != p2 {
			t.Error("Wrong loser")
		}
	}

	select {
	case <-time.After(500 * time.Millisecond):
		t.Error("expected game over response")
	case res := <-p2.Outgoing:
		if res.Type != GameOver {
			t.Errorf("EXpected %v, got %v", GameOver, res.Type)
		}
		payload := res.Payload.(GameOverPayload)
		if payload.Winner.player != p1 {
			t.Error("Wrong winner")
		}
		if payload.Loser.player != p2 {
			t.Error("Wrong loser")
		}
	}
}
