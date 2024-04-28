package request

import (
	"encoding/json"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/thosaphol/assessment-tax/utils"
)

type PersonalDeduction struct {
	Amount float64 `json:"amount" validate:"min=10000,max=100000.0" errormgs:"Invalid amount is required 10,000.0 to 100,000.0"`
}
type KReceiptDeduction struct {
	Amount float64 `json:"amount" validate:"min=0,max=100000.0" errormgs:"Invalid amount is required 0.0 to 100,000.0"`
}

func (d *PersonalDeduction) BindFromMap(m map[string]interface{}) error {
	jsonsTag := utils.GetJsonTags(*d)
	for _, jTag := range jsonsTag {
		_, ok := m[jTag]
		if !ok {
			return errors.New("Json structure invalid")
		}
	}

	// Convert the map to JSON
	jsonData, _ := json.Marshal(m)
	json.Unmarshal(jsonData, d)

	err := d.validate()
	if err != nil {
		return err
	}

	return nil

}

func (k *KReceiptDeduction) BindFromMap(m map[string]interface{}) error {
	jsonsTag := utils.GetJsonTags(*k)
	for _, jTag := range jsonsTag {
		_, ok := m[jTag]
		if !ok {
			return errors.New("Json structure invalid")
		}
	}

	// Convert the map to JSON
	jsonData, _ := json.Marshal(m)
	json.Unmarshal(jsonData, k)

	err := k.validate()
	if err != nil {
		return err
	}

	return nil

}

func (m *PersonalDeduction) validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return utils.ValidateFunc[PersonalDeduction](*m, validate, "errormgs")
}
func (k *KReceiptDeduction) validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return utils.ValidateFunc[KReceiptDeduction](*k, validate, "errormgs")
}
