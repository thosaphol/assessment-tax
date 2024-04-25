package repo

type Storer interface {
	SetPersonalDeduction(amount float64) error
}
