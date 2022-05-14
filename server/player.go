package server

import (
	"github.com/gorilla/websocket"
)

type Player struct {
	Name string

	Closing  chan bool
	Incoming chan string
	Outgoing chan string

	socket *websocket.Conn
}

func NewPlayer(socket *websocket.Conn) *Player {
	player := &Player{
		Closing:  make(chan bool),
		Incoming: make(chan string),
		Outgoing: make(chan string),

		socket: socket,
	}

	go player.Read()
	go player.Write()

	return player
}

func (p *Player) Send(message string) {
	p.Outgoing <- message
}

func (p *Player) Close() {
	p.Closing <- true
}

func (p *Player) Read() {
	for {
		_, msg, err := p.socket.ReadMessage()
		if err != nil {
			break
		}
		p.Incoming <- string(msg)
	}
}

func (p *Player) Write() {
	for {
		select {
		case msg := <-p.Outgoing:
			p.socket.WriteMessage(websocket.TextMessage, []byte(msg))
		case <-p.Closing:
			p.socket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		}
	}
}
