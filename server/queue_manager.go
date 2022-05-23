package server

type QueueManager struct {
	queue *Queue

	Unregister chan *Player
	Register   chan WithDispatcher
	Players    chan []*Player
}

func NewQueueManager() *QueueManager {
	manager := &QueueManager{
		queue: NewQueue(),

		Unregister: make(chan *Player),
		Register:   make(chan WithDispatcher),
		Players:    make(chan []*Player),
	}

	go func() {
		for {
			select {
			case player := <-manager.Unregister:
				manager.queue.Remove(player)

				player.Send(Response{
					Type: Dequeued,
				})
			case data := <-manager.Register:
				manager.queue.Queue(data.Player)

				data.Player.Send(Response{
					Type: WaitForMatch,
				})

				if manager.queue.Length() == 2 {
					players := make([]*Player, 0)

					for i := 0; i < 2; i++ {
						player := manager.queue.Dequeue()
						players = append(players, player)
					}

					data.Dispatcher.Dispatch <- Event{
						Type:    CreateMatch,
						Payload: players,
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
		qm.Register <- WithDispatcher{
			Player:     event.Player,
			Dispatcher: dispatcher,
		}
	case Dequeue:
		qm.Unregister <- event.Player
	}
}
