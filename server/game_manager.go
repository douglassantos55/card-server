package server

import (
	"log"
	"time"

	"github.com/google/uuid"
)

type Discarded struct {
	Cards  []string
	Player *Player
}

type Game struct {
	Id      uuid.UUID
	Players []*Player

	Ready   []*Player
	Current *Player
	Hands   map[*Player][]HasManaCost

	EndTurn chan *Player
	Discard chan Discarded
	Started chan time.Duration
}

func NewGame(players []*Player) *Game {
	game := &Game{
		Id:      uuid.New(),
		Players: players,

		Current: nil,
		Ready:   make([]*Player, 0),
		Hands:   make(map[*Player][]HasManaCost),

		EndTurn: make(chan *Player),
		Started: make(chan time.Duration),
		Discard: make(chan Discarded),
	}

	go func() {
		for {
			select {
			case <-game.Started:
				for _, player := range game.Players {
					hand := []HasManaCost{
						NewMinion(1, 1, 1),
						NewMinion(2, 1, 2),
						NewMinion(5, 4, 3),
					}

					game.Hands[player] = hand

					player.Send(Response{
						Type:    StartingHand,
						Payload: hand,
					})
				}
			case event := <-game.Discard:
				hand, ok := game.Hands[event.Player]

				if !ok {
					log.Printf("Hand not found for player %v, %v", hand, event.Player)
					return
				}

				for _, cardId := range event.Cards {
					for idx, card := range hand {
						if card.GetId() == cardId {
							game.Hands[event.Player] = append(
								game.Hands[event.Player][:idx],
								game.Hands[event.Player][idx+1:]...,
							)
						}
					}
				}

				for i := 0; i < len(event.Cards); i++ {
					game.Hands[event.Player] = append(
						game.Hands[event.Player],
						NewMinion(3, 1, 3),
					)
				}

				game.Ready = append(game.Ready, event.Player)

				event.Player.Send(Response{
					Type:    WaitOtherPlayers,
					Payload: game.Hands[event.Player],
				})

				if len(game.Ready) == len(game.Players) {
					game.StartTurns(75 * time.Second)
				}
			}
		}
	}()

	return game
}

func (g *Game) StartTurns(duration time.Duration) {
	g.Current = g.Players[1]
	g.NextTurn(duration)
}

func (g *Game) NextTurn(duration time.Duration) {
	var next *Player

	for _, player := range g.Players {
		if player == g.Current {
			player.Send(Response{
				Type: WaitTurn,
			})
		} else {
			next = player

			player.Send(Response{
				Type: StartTurn,
				Payload: TurnPayload{
					GameId:   g.Id,
					Duration: duration,
				},
			})
		}
	}
	g.Current = next

	go func() {
		select {
		case <-time.After(duration):
			g.NextTurn(duration)
			break
		case <-g.EndTurn:
			g.NextTurn(duration)
			break
		}
	}()

}

func (g *Game) Start(duration time.Duration) {
	go func() {
		select {
		case <-time.After(duration):
			g.StartTurns(75 * time.Second)
		}
	}()

	g.Started <- duration
}

func (g *Game) Process(event Event, dispatcher *Dispatcher) {
	switch event.Type {
	case CardsDiscarded:
		data := event.Payload.(CardsDiscardedPayload)
		uuid, err := uuid.Parse(data.GameId)

		if err != nil || uuid != g.Id {
			return
		}

		g.Discard <- Discarded{
			Cards:  data.Cards,
			Player: event.Player,
		}

	case EndTurn:
		uuid, err := uuid.Parse(event.Payload.(string))

		if err != nil || uuid != g.Id {
			return
		}

		g.EndTurn <- event.Player
	}
}

type GameManager struct{}

func NewGameManager() *GameManager {
	return &GameManager{}
}

func (gm *GameManager) Process(event Event, dispatcher *Dispatcher) {
	switch event.Type {
	case StartGame:
		players := event.Payload.([]*Player)
		game := NewGame(players)
		dispatcher.Register <- game
	}
}
