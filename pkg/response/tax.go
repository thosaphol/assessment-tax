package response

type Tax struct {
	Tax       float64    `json:"tax"`
	TaxLevels []TaxLevel `json:"taxLevel"`
}
type TaxLevel struct {
	Level string  `json:"level"`
	Tax   float64 `json:"tax"`
}
type TaxWithRefund struct {
	Tax
	TaxRefund float64 `json:"taxRefund"`
}

type TaxWithIncome struct {
	TotalIncome float64 `json:"totalIncome"`
	Tax         float64 `json:"tax"`
	TaxRefund   float64 `json:"taxRefund"`
}
type Taxes struct {
	Taxes []TaxWithIncome `json:"taxes"`
}
