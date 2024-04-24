package response

type Tax struct {
	Tax float64 `json:"tax"`
}
type TaxWithRefund struct {
	Tax
	TaxRefund float64 `json:"taxRefund"`
}
