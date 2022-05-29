package server

type Event struct {
	Type    EventType
	Player  *Player
	Payload interface{}
}

type EventType string

const (
	QueueUp         EventType = "queue_up"
	Dequeue         EventType = "dequeue"
	CreateMatch     EventType = "create_match"
	MatchConfirmed  EventType = "match_confirmed"
	MatchDeclined   EventType = "match_declined"
	AskConfirmation EventType = "confirm_match"
	StartGame       EventType = "start_game"
	CardsDiscarded  EventType = "cards_discarded"
	EndTurn         EventType = "end_turn"
	PlayCard        EventType = "play_card"
	Attack          EventType = "attack"
)

type CardsDiscardedPayload struct {
	GameId string
	Cards  []string
}

type PlayCardPayload struct {
	GameId string
	Card   string
}

type AttackPayload struct {
	GameId   string
	Attacker string
	Target   string
}
