package server

import "testing"

func TestReduceManaCost(t *testing.T) {
	card := NewMinion(2, 1, 1)

	card.ReduceManaCost(1)
	if card.GetManaCost() != 1 {
		t.Errorf("Expected %v, got %v", 1, card.GetManaCost())
	}

	card.ReduceManaCost(2)
	if card.GetManaCost() != 0 {
		t.Errorf("Expected %v, got %v", 0, card.GetManaCost())
	}
}

func TestMinionCardsHaveHealthAndDamage(t *testing.T) {
	card := NewMinion(2, 1, 3)

	if card.GetHealth() != 3 {
		t.Errorf("Expected %v, got %v", 3, card.GetHealth())
	}

	if card.GetDamage() != 1 {
		t.Errorf("Expected %v, got %v", 1, card.GetDamage())
	}
}

func TestReduceHealth(t *testing.T) {
	card := NewMinion(2, 1, 2)

	card.ReduceHealth(1)
	if card.GetHealth() != 1 {
		t.Errorf("Expected %v, got %v", 1, card.GetHealth())
	}

	card.ReduceHealth(2)
	if card.GetHealth() != 0 {
		t.Errorf("Expected %v, got %v", 0, card.GetHealth())
	}
}

func TestReduceDamage(t *testing.T) {
	card := NewMinion(2, 2, 3)

	card.ReduceDamage(1)
	if card.GetDamage() != 1 {
		t.Errorf("Expected %v, got %v", 1, card.GetDamage())
	}

	card.ReduceDamage(2)
	if card.GetDamage() != 0 {
		t.Errorf("Expected %v, got %v", 0, card.GetDamage())
	}
}

func TestGainHealth(t *testing.T) {
	card := NewMinion(2, 2, 3)

	card.GainHealth(1)
	if card.GetHealth() != 4 {
		t.Errorf("Expected %v, got %v", 4, card.GetHealth())
	}

	card.ReduceHealth(2)
	card.GainHealth(11)
	if card.GetHealth() != 13 {
		t.Errorf("Expected %v, got %v", 13, card.GetHealth())
	}
}

func TestGainDamage(t *testing.T) {
	card := NewMinion(2, 1, 1)

	card.GainDamage(1)
	if card.GetDamage() != 2 {
		t.Errorf("Expected %v, got %v", 2, card.GetDamage())
	}

	card.GainDamage(10)
	if card.GetDamage() != 12 {
		t.Errorf("Expected %v, got %v", 12, card.GetDamage())
	}
}

func TestAttack(t *testing.T) {
	offender := NewMinion(2, 1, 2)
	defender := NewMinion(2, 1, 2)

	offender.Attack(defender)

	if offender.GetHealth() != 1 {
		t.Errorf("Expected %v, got %v", 1, offender.GetHealth())
	}

	if defender.GetHealth() != 1 {
		t.Errorf("Expected %v, got %v", 1, defender.GetHealth())
	}
}
