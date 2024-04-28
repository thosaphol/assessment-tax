package tax

import (
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/thosaphol/assessment-tax/pkg/repo"
	"github.com/thosaphol/assessment-tax/pkg/request"
	resp "github.com/thosaphol/assessment-tax/pkg/response"
	"github.com/thosaphol/assessment-tax/utils"
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
		case "donation", "k-receipt":
			continue
		default:
			return c.JSON(http.StatusBadRequest, Err{Message: "AllowanceType is 'donation' or 'k-receipt' only"})
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

	alwKReceipt := 0.0
	alwDonate := 0.0
	for _, alw := range ie.Allowances {
		if alw.AllowanceType == "donation" {
			alwDonate += alw.Amount
			alwDonate = math.Min(100000.0, alwDonate)
		} else if alw.AllowanceType == "k-receipt" {
			alwKReceipt += alw.Amount
			alwKReceipt = math.Min(50000.0, alwKReceipt)
		}
	}
	alwTotal := alwKReceipt + alwDonate
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

func (h *Handler) CalculationCSV(c echo.Context) error {
	file, err := c.FormFile("taxFile")
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{err.Error()})
	}
	if ext := utils.GetFileExt(file.Filename); ext != ".csv" {
		return c.JSON(http.StatusBadRequest, Err{"File extension must is .csv"})
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	reader := utils.NewCsvReader(src)

	s := reader.ReadLine()
	if !s {
		_, err = reader.GetLine()
		return c.JSON(http.StatusInternalServerError, Err{err.Error()})
	}

	hRecord, _ := reader.GetLine()
	if len(hRecord) != 3 {
		return c.JSON(http.StatusInternalServerError, Err{"Header of content is 'totalIncome,wht,donation' only"})
	}
	if hRecord[0] != "totalIncome" || hRecord[1] != "wht" || hRecord[2] != "donation" {
		return c.JSON(http.StatusInternalServerError, Err{"Header of content is 'totalIncome,wht,donation' only"})
	}

	personalD, err := h.store.PersonalDeduction()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{err.Error()})
	}

	var taxes []resp.TaxWithIncome
	for reader.ReadLine() {
		rec, err := reader.GetLine()
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}

		if len(rec) != 3 {
			return c.JSON(http.StatusBadRequest, Err{"Some rows have columns not equal to 3."})
		}

		var taxInc resp.TaxWithIncome

		income, wht, donate, err := separateRecord(rec)
		if err != nil {
			return c.JSON(http.StatusBadRequest, Err{err.Error()})
		}

		iNet := incomeNet(income, donate, personalD)
		netTax, refund := calculateTaxNet(iNet, wht)

		taxInc = resp.TaxWithIncome{TotalIncome: income, Tax: netTax, TaxRefund: refund}

		taxes = append(taxes, taxInc)

	}

	return c.JSON(http.StatusOK, resp.Taxes{Taxes: taxes})
}

func incomeNet(income, totalAlw, PersonalDed float64) float64 {
	alwTotal := totalAlw

	alwTotal = math.Min(alwTotal, 100000.0)
	iNet := income - PersonalDed
	iNet -= alwTotal
	return iNet
}
func separateRecord(record []string) (float64, float64, float64, error) {
	if len(record) != 3 {
		return 0, 0, 0, errors.New("row has columns not equal to 3.")
	}
	income, err := strconv.ParseFloat(record[0], 64)
	if err != nil {
		return 0, 0, 0, errors.New("Income column has format incorrect")
	}

	wht, err := strconv.ParseFloat(record[1], 64)
	if err != nil {
		return 0, 0, 0, errors.New("Wht column has format incorrect")
	}

	donate, err := strconv.ParseFloat(record[2], 64)
	if err != nil {
		return 0, 0, 0, errors.New("Donate column has format incorrect")
	}
	return income, wht, donate, nil
}

func calculateTaxNet(incomeNet, wht float64) (tax float64, refund float64) {
	var tConsts = GetTaxConsts()
	ttax := 0.0
	for i := 0; i < len(tConsts); i++ {
		tConst := tConsts[i]
		if incomeNet > float64(tConst.Lower) {

			if incomeNet > float64(tConst.Upper) {
				taxInLevel := (tConst.Upper - tConst.Lower) * tConst.TaxRate / 100
				ttax += float64(taxInLevel)

			} else {
				diffLower := incomeNet - float64(tConst.Lower)
				tax := diffLower * float64(tConst.TaxRate) / 100
				ttax += tax

			}
		}
	}

	if ttax >= wht {
		return ttax - wht, 0
	}
	return 0, wht - ttax
}
