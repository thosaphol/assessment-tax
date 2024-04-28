package response

type PersonalDeduction struct {
	PersonalDeduction float64 `json:"personalDeduction"`
}

type KReceiptDeduction struct {
	KReceipt float64 `json:"kReceipt"`
}
