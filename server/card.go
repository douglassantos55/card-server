package server

import "github.com/google/uuid"

type HasManaCost interface {
	GetId() string
	GetManaCost() int
	ReduceManaCost(amount int)
	IncreaseManaCost(amount int)
}

type Minion interface {
	HasManaCost
	Defender
}

type Defender interface {
	HasDamage
	HasHealth
	CanCounterAttack() bool
}

type HasDamage interface {
	GetDamage() int
	GainDamage(amount int)
	ReduceDamage(amount int)
	Attack(defender Defender)
}

type HasHealth interface {
	GetHealth() int
	GainHealth(amount int)
	ReduceHealth(amount int)
}

type Card struct {
	Id       uuid.UUID
	ManaCost int
}

func (c *Card) GetId() string {
	return c.Id.String()
}

func (c *Card) GetManaCost() int {
	return c.ManaCost
}

func (c *Card) ReduceManaCost(amount int) {
	c.ManaCost -= amount
	if c.ManaCost < 0 {
		c.ManaCost = 0
	}
}

func (c *Card) IncreaseManaCost(amount int) {
	c.ManaCost += amount
}

type MinionCard struct {
	Card
	Damage int
	Health int
}

func NewMinion(manaCost, damage, health int) Minion {
	return &MinionCard{
		Card: Card{
			Id:       uuid.New(),
			ManaCost: manaCost,
		},
		Damage: damage,
		Health: health,
	}
}

func (m *MinionCard) GetDamage() int {
	return m.Damage
}

func (m *MinionCard) GainDamage(amount int) {
	m.Damage += amount
}

func (m *MinionCard) ReduceDamage(amount int) {
	m.Damage -= amount
	if m.Damage < 0 {
		m.Damage = 0
	}
}

func (m *MinionCard) Attack(defender Defender) {
	defender.ReduceHealth(m.GetDamage())

	if defender.CanCounterAttack() {
		m.ReduceHealth(defender.GetDamage())
	}
}

func (m *MinionCard) GetHealth() int {
	return m.Health
}

func (m *MinionCard) GainHealth(amount int) {
	m.Health += amount
}

func (m *MinionCard) ReduceHealth(amount int) {
	m.Health -= amount
	if m.Health < 0 {
		m.Health = 0
	}
}

func (m *MinionCard) CanCounterAttack() bool {
	return true
}
