package deduction

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/thosaphol/assessment-tax/pkg/repo"
	"github.com/thosaphol/assessment-tax/pkg/request"
	"github.com/thosaphol/assessment-tax/pkg/response"
)

// import (
// 	"github.com/labstack/echo"
// 	"github.com/thosaphol/assessment-tax/pkg/tax"
// )

type Handler struct {
	store repo.Storer
}

func New(db repo.Storer) *Handler {
	return &Handler{store: db}
}

func (h *Handler) SetDeductionPersonal(c echo.Context) error {
	var reqD request.Deduction
	err := c.Bind(&reqD)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Err{Message: err.Error()})
	}

	err = h.store.SetPersonalDeduction(reqD.Amount)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Err{Message: err.Error()})
	}

	var resp = response.PersonalDeduction{PersonalDeduction: reqD.Amount}
	return c.JSON(http.StatusOK, resp)
}
