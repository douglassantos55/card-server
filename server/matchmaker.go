package server

import "time"

type Matchmaker struct{}

func NewMatchmaker() *Matchmaker {
	return &Matchmaker{}
}

func (m *Matchmaker) Process(event Event, dispatcher *Dispatcher) {
	switch event.Type {
	case CreateMatch:
		players := event.Payload.([]*Player)
		match := NewMatch(players, 5*time.Second)

		dispatcher.Register <- match

		dispatcher.Dispatch <- Event{
			Type:    AskConfirmation,
			Payload: match.Id,
		}
	}
}
