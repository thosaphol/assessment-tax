package repo

type Storer interface {
	SetPersonalDeduction(amount float64) error
	PersonalDeduction() (float64, error)
}
