package server

type Event struct {
	Type   EventType
	Player *Player
}

type EventType string

const (
	QueueUp EventType = "queue_up"
	Dequeue EventType = "dequeue"
)
