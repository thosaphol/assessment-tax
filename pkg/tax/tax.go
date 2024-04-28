package tax

import "math"

type TaxConst struct {
	Lower   float64
	Upper   float64
	TaxRate int
	Level   string
}

func GetTaxConsts() []TaxConst {
	return []TaxConst{
		{Lower: 0, Upper: 150000, TaxRate: 0, Level: "0-150,000"},
		{Lower: 150000, Upper: 500000, TaxRate: 10, Level: "150,001-500,000"},
		{Lower: 500000, Upper: 1000000, TaxRate: 15, Level: "500,001-1,000,000"},
		{Lower: 1000000, Upper: 2000000, TaxRate: 20, Level: "1,000,001-2,000,000"},
		{Lower: 2000000, Upper: math.MaxInt, TaxRate: 35, Level: "2,000,001 ขึ้นไป"},
	}
}
