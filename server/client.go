package server

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	Incoming chan string
}

func NewClient(addr string) *Client {
	client := &Client{
		Incoming: make(chan string),
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
			_, message, err := socket.ReadMessage()

			if err != nil {
                break
			}

			if message != nil {
				c.Incoming <- string(message)
			}
		}

        c.Incoming <- "done"
	}()
}
