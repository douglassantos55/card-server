package server

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type Deck struct {
	cards []HasManaCost
}

func (d *Deck) Draw() HasManaCost {
	card := d.cards[0]
	d.cards = append(d.cards[1:])
	return card
}

func (d *Deck) Count() int {
	return len(d.cards)
}

func (d *Deck) DrawMany(count int) []HasManaCost {
	var cards []HasManaCost
	for i := 0; i < count; i++ {
		cards = append(cards, d.Draw())
	}
	return cards
}

func (d *Deck) Add(card HasManaCost) {
	d.cards = append(d.cards, card)
}

func NewDeck() *Deck {
	rand.Seed(time.Now().UnixNano())

	var cards []HasManaCost
	for i := 0; i < 60; i++ {
		cards = append(
			cards,
			NewMinion(rand.Intn(10), rand.Intn(10), rand.Intn(10)),
		)
	}
	return &Deck{cards: cards}
}

type Discarded struct {
	Cards  []string
	Player *Player
}

type GamePlayer struct {
	player *Player

	Deck    *Deck
	Health  int
	Mana    int
	MaxMana int
	Board   *Board
	Hand    []HasManaCost

	Current bool
}

func (gp *GamePlayer) GetHealth() int {
	return gp.Health
}
func (gp *GamePlayer) GainHealth(amount int) {
	gp.Health += amount
}
func (gp *GamePlayer) ReduceHealth(amount int) {
	gp.Health -= amount
}

func (gp *GamePlayer) Send(response Response) {
	gp.player.Send(response)
}

func (gp *GamePlayer) IncreaseMana(amount int) {
	gp.MaxMana += amount
	if gp.MaxMana > 10 {
		gp.MaxMana = 10
	}
}

func (gp *GamePlayer) GainMana(amount int) {
	gp.Mana += amount
	if gp.Mana > gp.MaxMana {
		gp.Mana = gp.MaxMana
	}
}

func (gp *GamePlayer) RefillMana() {
	gp.Mana = gp.MaxMana
}

func (gp *GamePlayer) ConsumeMana(amount int) {
	gp.Mana -= amount
	if gp.Mana < 0 {
		gp.Mana = 0
	}
}

type Game struct {
	Id      uuid.UUID
	Ready   []*Player
	Players map[*Player]*GamePlayer

	EndTurn   chan *Player
	Discard   chan Discarded
	Started   chan time.Duration
	StartTurn chan time.Duration
	TurnOver  chan time.Duration
	PlayCard  chan PlayCardPayload
	Attack    chan AttackPayload
}

func NewGame(players []*Player) *Game {
	gamePlayers := map[*Player]*GamePlayer{}

	for idx, player := range players {
		deck := NewDeck()

		gamePlayers[player] = &GamePlayer{
			player:  player,
			Health:  30,
			Deck:    deck,
			Current: idx == 0,
			Board:   NewBoard(),
			Hand:    deck.DrawMany(3),
		}
	}

	game := &Game{
		Id:      uuid.New(),
		Ready:   make([]*Player, 0),
		Players: gamePlayers,

		EndTurn:   make(chan *Player),
		Started:   make(chan time.Duration),
		Discard:   make(chan Discarded),
		StartTurn: make(chan time.Duration),
		TurnOver:  make(chan time.Duration),
		PlayCard:  make(chan PlayCardPayload),
		Attack:    make(chan AttackPayload),
	}

	go func() {
		for {
			select {
			case <-game.Started:
				for _, player := range game.Players {
					go player.Send(Response{
						Type:    StartingHand,
						Payload: player.Hand,
					})
				}
			case event := <-game.Discard:
				player, ok := game.Players[event.Player]

				if !ok {
					return
				}

				for _, cardId := range event.Cards {
					for idx, card := range player.Hand {
						if card.GetId() == cardId {
							player.Hand = append(
								player.Hand[:idx],
								player.Hand[idx+1:]...,
							)

							player.Deck.Add(card)
						}
					}
				}

				player.Hand = append(
					player.Hand,
					player.Deck.DrawMany(len(event.Cards))...,
				)

				game.Ready = append(game.Ready, event.Player)

				player.Send(Response{
					Type:    WaitOtherPlayers,
					Payload: player.Hand,
				})

				if len(game.Ready) == len(game.Players) {
					go game.StartTurns(75 * time.Second)
				}
			case duration := <-game.StartTurn:
				for _, player := range game.Players {
					if player.Current {
						player.IncreaseMana(1)
						player.RefillMana()

						card := player.Deck.Draw()
						player.Hand = append(player.Hand, card)

						for _, minion := range player.Board.Defenders {
							minion.SetStatus(&Ready{})
						}

						go player.Send(Response{
							Type: StartTurn,
							Payload: TurnPayload{
								GameId:      game.Id,
								CardsLeft:   player.Deck.Count(),
								Card:        card,
								CardsInHand: len(player.Hand),
								Mana:        player.Mana,
								Duration:    duration,
							},
						})
					} else {
						go player.Send(Response{
							Type: WaitTurn,
						})
					}
				}

				go func() {
					select {
					case <-time.After(duration):
						game.TurnOver <- duration
					case <-game.EndTurn:
						game.TurnOver <- duration
					}
				}()
			case duration := <-game.TurnOver:
				for _, player := range game.Players {
					player.Current = !player.Current
				}
				go game.StartTurns(duration)
			case data := <-game.PlayCard:
				var index int
				var card HasManaCost
				var current *GamePlayer

				for _, player := range game.Players {
					for idx, c := range player.Hand {
						if c.GetId() == data.Card {
							card = c
							index = idx

						}
					}
					if player.Current {
						current = player
					}
				}

				if card == nil {
					current.Send(Response{
						Type:    Error,
						Payload: "Card not found",
					})
				} else if card.GetManaCost() > current.Mana {
					current.Send(Response{
						Type:    Error,
						Payload: "Not enough mana",
					})
				} else {
					current.Hand = append(
						current.Hand[:index],
						current.Hand[index+1:]...,
					)

					current.ConsumeMana(card.GetManaCost())
					played := current.Board.PlaceCard(card.(Defender))

					for _, player := range game.Players {
						go player.Send(Response{
							Type:    CardPlayed,
							Payload: played,
						})
					}
				}
			case data := <-game.Attack:
				var current *GamePlayer
				var other *GamePlayer

				for _, player := range game.Players {
					if player.Current {
						current = player
					} else {
						other = player
					}
				}

				var attacker ActiveDefender
				var defender ActiveDefender

				for _, card := range current.Board.Defenders {
					if card.GetId() == data.Attacker {
						attacker = card.(ActiveDefender)
						break
					}
				}

				if data.Target != "" {
					for _, card := range other.Board.Defenders {
						if card.GetId() == data.Target {
							defender = card.(ActiveDefender)
							break
						}
					}

					if attacker.CanAttack() {
						attacker.Attack(defender)

						if attacker.GetHealth() == 0 {
							current.Board.Remove(attacker)
						}
						if defender.GetHealth() == 0 {
							other.Board.Remove(defender)
						}

						current.Send(Response{
							Type: AttackResult,
							Payload: []*Board{
								current.Board,
								other.Board,
							},
						})

						other.Send(Response{
							Type: AttackResult,
							Payload: []*Board{
								other.Board,
								current.Board,
							},
						})
					} else {
						current.Send(Response{
							Type:    Error,
							Payload: "Cannot attack with this card",
						})
					}
				} else if len(other.Board.Defenders) == 0 {
					other.ReduceHealth(attacker.GetDamage())

					for _, player := range game.Players {
						player.Send(Response{
							Type: DamageTaken,
							Payload: DamageTakenPayload{
								Health: other.GetHealth(),
							},
						})
					}
				} else {
					current.Send(Response{
						Type:    Error,
						Payload: "Cannot attack player with minions on board",
					})
				}
			}
		}
	}()

	return game
}

func (g *Game) StartTurns(duration time.Duration) {
	g.StartTurn <- duration
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
	case PlayCard:
		data := event.Payload.(PlayCardPayload)
		uuid, err := uuid.Parse(data.GameId)

		if err != nil || uuid != g.Id {
			return
		}

		g.PlayCard <- data
	case Attack, AttackPlayer:
		data := event.Payload.(AttackPayload)
		uuid, err := uuid.Parse(data.GameId)

		if err != nil || uuid != g.Id {
			return
		}

		g.Attack <- data
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
