package server

import (
	"time"

	"github.com/google/uuid"
)

type WithDispatcher struct {
	Player     *Player
	Dispatcher *Dispatcher
}

type Match struct {
	Id        uuid.UUID
	Players   []*Player
	Confirmed []*Player
	Duration  time.Duration

	Ready   chan bool
	Confirm chan WithDispatcher
	Cancel  chan *Dispatcher
}

func NewMatch(players []*Player, confirmDuration time.Duration) *Match {
	match := &Match{
		Id:        uuid.New(),
		Players:   players,
		Duration:  confirmDuration,
		Confirmed: make([]*Player, 0),

		Ready:   make(chan bool),
		Confirm: make(chan WithDispatcher),
		Cancel:  make(chan *Dispatcher),
	}

	go func() {
		for {
			select {
			case dispatcher := <-match.Cancel:
				for _, player := range match.Players {
					player.Send(Response{
						Type: MatchCanceled,
					})
				}
				for _, player := range match.Confirmed {
					dispatcher.Dispatch <- Event{
						Type:   QueueUp,
						Player: player,
					}
				}

				dispatcher.Unregister <- match
			case data := <-match.Confirm:
				data.Player.Send(Response{
					Type: WaitOtherPlayers,
				})

				match.Confirmed = append(match.Confirmed, data.Player)

				if len(match.Confirmed) == len(match.Players) {
					data.Dispatcher.Dispatch <- Event{
						Type: StartGame,
					}

					match.Ready <- true
					data.Dispatcher.Unregister <- match
				}
			}
		}
	}()

	return match
}

func (m *Match) Process(event Event, dispatcher *Dispatcher) {
	switch event.Type {
	case AskConfirmation:
		uuid := event.Payload.(uuid.UUID)

		if uuid != m.Id {
			return
		}

		for _, player := range m.Players {
			go player.Send(Response{
				Type:    MatchFound,
				Payload: m.Id,
			})
		}

		go func() {
			select {
			case <-time.After(m.Duration):
				dispatcher.Dispatch <- Event{
					Type:    MatchDeclined,
					Payload: m.Id.String(),
				}
			case <-m.Ready:
				break
			}
		}()
	case MatchConfirmed:
		uuid, err := uuid.Parse(event.Payload.(string))

		if err != nil || uuid != m.Id {
			return
		}

		for _, player := range m.Players {
			if player == event.Player {
				m.Confirm <- WithDispatcher{
					Player:     player,
					Dispatcher: dispatcher,
				}
			}
		}
	case MatchDeclined:
		uuid, err := uuid.Parse(event.Payload.(string))
		if err != nil || uuid != m.Id {
			return
		}
		m.Cancel <- dispatcher
	}
}
