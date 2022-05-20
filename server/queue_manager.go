package server

type QueueManager struct {
	queue *Queue

	Register   chan *Player
	Unregister chan *Player
}

func NewQueueManager() *QueueManager {
	manager := &QueueManager{
		queue: NewQueue(),

		Register:   make(chan *Player),
		Unregister: make(chan *Player),
	}

	go func() {
		for {
			select {
			case player := <-manager.Unregister:
				manager.queue.Remove(player)

				player.Send(Response{
					Type: Dequeued,
				})
			case player := <-manager.Register:
				manager.queue.Queue(player)

				player.Send(Response{
					Type: WaitForMatch,
				})

				if manager.queue.Length() == 2 {
					for i := 0; i < 2; i++ {
						player := manager.queue.Dequeue()

						go player.Send(Response{
							Type: MatchFound,
						})
					}
				}
			}
		}
	}()

	return manager
}

func (qm *QueueManager) Process(event Event, dispatcher *Dispatcher) {
	switch event.Type {
	case QueueUp:
		qm.Register <- event.Player
	case Dequeue:
		qm.Unregister <- event.Player
	}
}
