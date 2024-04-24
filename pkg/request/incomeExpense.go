package request

type IncomeExpense struct {
	TotalIncome float64      `json:"totalIncome"`
	Wht         float64      `json:"wht"`
	Allowances  []Allowances `json:"allowances"`
}

type Allowances struct {
	AllowanceType string  `json:"allowanceType"`
	Amount        float64 `json:"amount"`
}
