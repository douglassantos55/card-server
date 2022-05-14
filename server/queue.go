package server

type node struct {
	Next   *node
	Player *Player
}

type Queue struct {
	head    *node
	players map[*Player]*node
}

func NewQueue() *Queue {
	return &Queue{
		head:    nil,
		players: make(map[*Player]*node),
	}
}

func (q *Queue) Queue(player *Player) {
	node := &node{Player: player}

	if q.head == nil {
		q.head = node
		q.players[player] = nil
	} else {
		cur := q.head
		for cur.Next != nil {
			cur = cur.Next
		}
		cur.Next = node
		q.players[player] = cur
	}
}

func (q *Queue) Dequeue() *Player {
	node := q.head
	if q.head != nil {
		q.head = q.head.Next
	}
	if node == nil {
		return nil
	}
	return node.Player
}

func (q *Queue) Remove(player *Player) bool {
	prev, ok := q.players[player]

	if !ok {
		return false
	}

	if prev != nil {
		prev.Next = prev.Next.Next
	} else if q.head.Next != nil {
		q.head = q.head.Next
	} else {
		q.head = nil
	}

	delete(q.players, player)
	return true
}
