package server

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	Incoming chan Response // incoming from server
	Outgoing chan Event    // outgoing to server
}

func NewClient(addr string) *Client {
	client := &Client{
		Incoming: make(chan Response),
		Outgoing: make(chan Event),
	}

	client.Connect(addr)
	return client
}

func (c *Client) Connect(addr string) {
	socket, _, err := websocket.DefaultDialer.Dial("ws://"+addr, nil)

	if err != nil {
		log.Printf("Could not connect to server at %v\n", "ws://"+addr)
		return
	}

	go func() {
		defer socket.Close()

		for {
			var response Response
			err := socket.ReadJSON(&response)

			if err != nil {
				break
			}

			c.Incoming <- response
		}
	}()

	go func() {
		for {
			select {
			case event := <-c.Outgoing:
				socket.WriteJSON(event)
			}
		}
	}()
}
