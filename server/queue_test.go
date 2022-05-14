package server

import "testing"

func TestDequeueFirstInQueue(t *testing.T) {
	queue := NewQueue()
	expected := &Player{Name: "test"}

	queue.Queue(expected)
	queue.Queue(&Player{Name: "other"})

	got := queue.Dequeue()

	if got != expected {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestDequeueEmpty(t *testing.T) {
	queue := NewQueue()
	got := queue.Dequeue()

	if got != nil {
		t.Errorf("Expected %v, got %v", nil, got)
	}
}

func TestRemovePlayer(t *testing.T) {
	queue := NewQueue()

	player := &Player{Name: "test"}
	expected := &Player{Name: "other"}

	queue.Queue(player)
	queue.Queue(expected)
	queue.Queue(&Player{Name: "another"})

	if !queue.Remove(player) {
		t.Error("Expected player to be removed")
	}

	got := queue.Dequeue()

	if got != expected {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestRemoveEmpty(t *testing.T) {
	queue := NewQueue()
	if queue.Remove(&Player{}) {
		t.Error("Should not remove from empty queue")
	}
}

func TestRemoveEmptied(t *testing.T) {
	queue := NewQueue()

	queue.Queue(&Player{})
	queue.Queue(&Player{})
	queue.Queue(&Player{})

	queue.Dequeue()
	queue.Dequeue()
	queue.Dequeue()

	if queue.Dequeue() != nil {
		t.Error("Expected queue to be empty")
	}

	if queue.Remove(&Player{}) {
		t.Error("Should not remove from empty queue")
	}
}
