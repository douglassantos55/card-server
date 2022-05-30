package main

import "example.com/wingscam-server/server"

func main() {
    dispatcher := server.NewDispatcher()

    dispatcher.Register <- server.NewQueueManager()
    dispatcher.Register <- server.NewMatchmaker()
    dispatcher.Register <- server.NewGameManager()

	server := server.NewServer(dispatcher)
	server.Listen("0.0.0.0:8080")
}
