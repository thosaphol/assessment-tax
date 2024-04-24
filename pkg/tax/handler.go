package tax

import (
	"math"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/thosaphol/assessment-tax/pkg/request"
	resp "github.com/thosaphol/assessment-tax/pkg/response"
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
	err := c.Bind(&ie)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{err.Error()})
	}

	for _, alw := range ie.Allowances {
		if alw.Amount < 0 {
			return c.JSON(http.StatusBadRequest, Err{Message: "Amount allowance must greater than 0."})
		}
		switch alw.AllowanceType {
		case "donation":
			continue
		default:
			return c.JSON(http.StatusBadRequest, Err{Message: "AllowanceType is 'donation' only"})
		}
	}

	if ie.TotalIncome < 0 {
		return c.JSON(http.StatusBadRequest, Err{"TotalIncome must have a starting value of 0."})
	}

	if ie.Wht < 0 || ie.Wht > ie.TotalIncome {
		return c.JSON(http.StatusBadRequest, Err{"Wht must be in the range 0 to TotalIncome."})
	}

	var taxLevels = GetTaxConsts()

	alwTotal := 0.0
	for _, alw := range ie.Allowances {
		alwTotal += alw.Amount
	}
	alwTotal = math.Min(alwTotal, 100000.0)
	iNet := ie.TotalIncome - 60000
	iNet -= alwTotal

	ttax := 0.0

	for i := 0; i < len(taxLevels); i++ {
		taxLevel := taxLevels[i]

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

	if ttax >= ie.Wht {
		return c.JSON(http.StatusOK, resp.Tax{Tax: ttax - ie.Wht})
	}
	return c.JSON(http.StatusOK, resp.TaxWithRefund{Tax: resp.Tax{0}, TaxRefund: ie.Wht - ttax})
}
