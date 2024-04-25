package tax

import (
	"math"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/thosaphol/assessment-tax/pkg/repo"
	"github.com/thosaphol/assessment-tax/pkg/request"
	resp "github.com/thosaphol/assessment-tax/pkg/response"
)

type Err struct {
	Message string `json:"message"`
}

type Handler struct {
	store repo.Storer
}

func New(db repo.Storer) *Handler {
	return &Handler{store: db}
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

	personalD, err := h.store.PersonalDeduction()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{err.Error()})
	}

	var tConsts = GetTaxConsts()
	alwTotal := 0.0
	for _, alw := range ie.Allowances {
		alwTotal += alw.Amount
	}
	alwTotal = math.Min(alwTotal, 100000.0)
	iNet := ie.TotalIncome - personalD
	iNet -= alwTotal

	var tLevels []resp.TaxLevel
	ttax := 0.0
	for i := 0; i < len(tConsts); i++ {
		tConst := tConsts[i]
		var tLevel = resp.TaxLevel{Level: tConst.Level}
		if iNet > float64(tConst.Lower) {

			if iNet > float64(tConst.Upper) {
				taxInLevel := (tConst.Upper - tConst.Lower) * tConst.TaxRate / 100
				ttax += float64(taxInLevel)

				tLevel.Tax = float64(taxInLevel)
			} else {
				diffLower := iNet - float64(tConst.Lower)
				tax := diffLower * float64(tConst.TaxRate) / 100
				ttax += tax

				tLevel.Tax = float64(tax)
			}
		}
		tLevels = append(tLevels, tLevel)
	}

	if ttax >= ie.Wht {
		return c.JSON(http.StatusOK, resp.Tax{Tax: ttax - ie.Wht, TaxLevels: tLevels})
	}
	return c.JSON(http.StatusOK, resp.TaxWithRefund{Tax: resp.Tax{Tax: 0, TaxLevels: tLevels},
		TaxRefund: ie.Wht - ttax})
}
