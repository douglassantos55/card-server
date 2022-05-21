package server

import "time"

type QueueManager struct {
	queue *Queue

	Unregister chan *Player
	Register   chan *Player
	Players    chan []*Player
}

func NewQueueManager() *QueueManager {
	manager := &QueueManager{
		queue: NewQueue(),

		Unregister: make(chan *Player),
		Register:   make(chan *Player),
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
			case player := <-manager.Register:
				manager.queue.Queue(player)

				player.Send(Response{
					Type: WaitForMatch,
				})

				if manager.queue.Length() == 2 {
					players := make([]*Player, 0)
					for i := 0; i < 2; i++ {
						player := manager.queue.Dequeue()
						players = append(players, player)

						go player.Send(Response{
							Type: MatchFound,
						})
					}

					manager.Players <- players
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

		select {
		case players := <-qm.Players:
			dispatcher.Dispatch <- Event{
				Type:    CreateMatch,
				Payload: players,
			}
		case <-time.After(time.Millisecond):
		}
	case Dequeue:
		qm.Unregister <- event.Player
	}
}
