package server

type Board struct {
	Defenders []ActiveDefender
}

func NewBoard() *Board {
	return &Board{
		Defenders: make([]ActiveDefender, 0),
	}
}

type Status interface {
	CanAttack() bool
	CanCounterAttack() bool
}

type HasStatus interface {
	Status
	GetStatus() Status
	SetStatus(status Status)
}

type ActiveDefender interface {
	Defender
	HasStatus
	Attack(defender ActiveDefender)
}

type Exhausted struct{}

func (t *Exhausted) CanAttack() bool {
	return false
}

func (t *Exhausted) CanCounterAttack() bool {
	return true
}

type Ready struct{}

func (r *Ready) CanAttack() bool {
	return true
}

func (r *Ready) CanCounterAttack() bool {
	return true
}

type Frozen struct{}

func (f *Frozen) CanAttack() bool {
	return false
}

func (f *Frozen) CanCounterAttack() bool {
	return true
}

type ActiveMinion struct {
	Defender
	Status Status
}

func (t *ActiveMinion) GetStatus() Status {
	return t.Status
}

func (t *ActiveMinion) SetStatus(status Status) {
	t.Status = status
}

func (t *ActiveMinion) CanAttack() bool {
	return t.Status.CanAttack()
}

func (t *ActiveMinion) CanCounterAttack() bool {
	return t.Status.CanCounterAttack()
}

func (t *ActiveMinion) Attack(defender ActiveDefender) {
	if !t.CanAttack() {
		panic("Attacking when should not be able to")
	}

	defender.ReduceHealth(t.GetDamage())

	if defender.CanCounterAttack() {
		t.ReduceHealth(defender.GetDamage())
	}

	t.SetStatus(&Exhausted{})
}

func (b *Board) PlaceCard(card Defender) ActiveDefender {
	if len(b.Defenders) >= 7 {
		return nil
	}

	defender := &ActiveMinion{
		Defender: card,
		Status:   &Exhausted{},
	}

	b.Defenders = append(b.Defenders, defender)
	return defender
}
