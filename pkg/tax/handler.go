package tax

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/thosaphol/assessment-tax/pkg/request"
	"github.com/thosaphol/assessment-tax/pkg/response"
)

type Err struct {
	Message string `json:"message"`
}

type Handler struct {
}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) Calculation(c echo.Context) error {

	var ie request.IncomeExpense
	_ = c.Bind(&ie)

	var taxLevels = GetTaxConsts()

	ttax := 0.0

	for i := 0; i < len(taxLevels); i++ {
		taxLevel := taxLevels[i]
		iNet := ie.TotalIncome - 60000

		if iNet > float64(taxLevel.Lower) {

			if iNet > float64(taxLevel.Upper) {
				taxInLevel := (taxLevel.Upper - taxLevel.Lower) * taxLevel.TaxRate / 100
				ttax += float64(taxInLevel)
			} else {
				diffLower := iNet - float64(taxLevel.Lower)
				tax := diffLower * float64(taxLevel.TaxRate) / 100
				ttax += tax
			}

		}
	}

	return c.JSON(http.StatusOK, response.Tax{Tax: ttax})
}
