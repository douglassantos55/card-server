package server

import "testing"

func TestMinionsStartExhausted(t *testing.T) {
	board := NewBoard()
	card := NewMinion(2, 1, 1)

	decorated := board.PlaceCard(card)

	if decorated.CanAttack() {
		t.Error("expected card to start exhausted")
	}
}

func TestCanPlaceUpTo7MinionsOnBoard(t *testing.T) {
	board := NewBoard()

	board.PlaceCard(NewMinion(1, 1, 1))
	board.PlaceCard(NewMinion(1, 1, 1))
	board.PlaceCard(NewMinion(1, 1, 1))
	board.PlaceCard(NewMinion(1, 1, 1))
	board.PlaceCard(NewMinion(1, 1, 1))
	board.PlaceCard(NewMinion(1, 1, 1))
	board.PlaceCard(NewMinion(1, 1, 1))

	if board.PlaceCard(NewMinion(1, 1, 1)) != nil {
		t.Errorf("Should not place more than 7 cards on board")
	}
}

func TestCanAttackEnemies(t *testing.T) {
	attacker := &ActiveMinion{
		Defender: NewMinion(1, 2, 2),
		Status:   &Ready{},
	}

	defender := &ActiveMinion{
		Defender: NewMinion(1, 1, 1),
		Status:   &Ready{},
	}

	attacker.Attack(defender)
	if defender.GetHealth() != 0 {
		t.Errorf("Expected %v, got %v", 0, defender.GetHealth())
	}

	if attacker.GetHealth() != 2 {
		t.Errorf("Expected %v, got %v", 2, attacker.GetHealth())
	}
}

func TestAttackingExhaustsMinion(t *testing.T) {
	attacker := &ActiveMinion{
		Defender: NewMinion(1, 2, 2),
		Status:   &Ready{},
	}

	defender := &ActiveMinion{
		Defender: NewMinion(1, 1, 1),
		Status:   &Ready{},
	}

	attacker.Attack(defender)

	if attacker.CanAttack() {
		t.Error("Expected minion to be exhausted after attacking")
	}
}

func TestFrozenCannotAttack(t *testing.T) {
	attacker := &ActiveMinion{
		Defender: NewMinion(1, 2, 2),
		Status:   &Frozen{},
	}

	if attacker.CanAttack() {
		t.Error("Frozen should not be able to attack")
	}

	if !attacker.CanCounterAttack() {
		t.Error("Frozen should be able to counter-attack")
	}
}
