package server

type GameManager struct{}

type Game struct {
	Players []*Player
}

func NewGame(players []*Player) *Game {
	return &Game{
		Players: players,
	}
}

func (g *Game) Process(event Event, dispatcher *Dispatcher) {
}

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
