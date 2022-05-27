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

	Ready []*Player
	Hands map[*Player][]HasManaCost

	Discard chan Discarded
	Started chan time.Duration
}

func NewGame(players []*Player) *Game {
	game := &Game{
		Id:      uuid.New(),
		Players: players,

		Ready: make([]*Player, 0),
		Hands: make(map[*Player][]HasManaCost),

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
					game.StartTurns()
				}
			}
		}
	}()

	return game
}

func (g *Game) StartTurns() {
	g.Players[0].Send(Response{
		Type: StartTurn,
	})

	g.Players[1].Send(Response{
		Type: WaitTurn,
	})
}

func (g *Game) Start(duration time.Duration) {
	go func() {
		select {
		case <-time.After(duration):
			g.StartTurns()
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
