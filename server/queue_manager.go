package server

type QueueManager struct {
	queue    *Queue
	Register chan *Player
}

func NewQueueManager() *QueueManager {
	manager := &QueueManager{
		queue:    NewQueue(),
		Register: make(chan *Player),
	}

	go func() {
		for {
			select {
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
	}
}
