package server

import (
	"time"

	"github.com/google/uuid"
)

type Response struct {
	Type    ResponseType
	Payload interface{}
}

type ResponseType string

const (
	Welcome          ResponseType = "welcome"
	WaitForMatch     ResponseType = "wait_for_match"
	MatchFound       ResponseType = "match_found"
	Dequeued         ResponseType = "dequeued"
	WaitOtherPlayers ResponseType = "wait_other_players"
	MatchCanceled    ResponseType = "match_canceled"
	StartingHand     ResponseType = "starting_hand"
	StartTurn        ResponseType = "start_turn"
	WaitTurn         ResponseType = "wait_turn"
	CardPlayed       ResponseType = "card_played"

	Error ResponseType = "error"
)

type TurnPayload struct {
	GameId      uuid.UUID
	Duration    time.Duration
	Card        HasManaCost
	Mana        int
	CardsLeft   int
	CardsInHand int
}
