package server

import (
	"github.com/gorilla/websocket"
)

type Player struct {
	Name string

	Closing  chan bool
	Incoming chan Event
	Outgoing chan Response

	socket *websocket.Conn
}

func NewPlayer(socket *websocket.Conn) *Player {
	player := &Player{
		Closing:  make(chan bool),
		Incoming: make(chan Event),
		Outgoing: make(chan Response),

		socket: socket,
	}

	go player.Read()
	go player.Write()

	return player
}

func (p *Player) Send(response Response) {
	p.Outgoing <- response
}

func (p *Player) Close() {
	p.Closing <- true
}

func (p *Player) Read() {
	for {
		var event Event
		err := p.socket.ReadJSON(&event)
		if err != nil {
			break
		}
		p.Incoming <- event
	}
}

func (p *Player) Write() {
	for {
		select {
		case msg := <-p.Outgoing:
			p.socket.WriteJSON(msg)
		case <-p.Closing:
			p.socket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		}
	}
}
