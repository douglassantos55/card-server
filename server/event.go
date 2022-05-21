package server

type Event struct {
	Type    EventType
	Player  *Player
	Payload interface{}
}

type EventType string

const (
	QueueUp     EventType = "queue_up"
	Dequeue     EventType = "dequeue"
	CreateMatch EventType = "create_match"
)
