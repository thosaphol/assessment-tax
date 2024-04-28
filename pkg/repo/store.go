package repo

type Storer interface {
	SetPersonalDeduction(amount float64) error
	PersonalDeduction() (float64, error)
	SetKReceiptDeduction(amount float64) error
	KReceiptDeduction() (float64, error)
}
