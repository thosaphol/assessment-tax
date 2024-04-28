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

	err = validateIncomeExpense(ie)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{err.Error()})
	}
	personalD, err := h.store.PersonalDeduction()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{err.Error()})
	}

	var tConsts = GetTaxConsts()

	alwTotal := calculateAllowance(ie.Allowances)
	iNet := calculateIncome(ie.TotalIncome, alwTotal, personalD)

	var tLevels []resp.TaxLevel
	ttax := 0.0
	for _, tConst := range tConsts {
		var tLevel = resp.TaxLevel{Level: tConst.Level}
		if iNet > float64(tConst.Lower) {

			if iNet > tConst.Upper {
				taxInLevel := calculateTaxLevel(tConst.Upper, tConst)
				ttax += taxInLevel

				tLevel.Tax = taxInLevel
			} else {
				taxInLevel := calculateTaxLevel(iNet, tConst)
				ttax += taxInLevel

				tLevel.Tax = taxInLevel
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

func calculateAllowance(alws []request.Allowance) float64 {
	alwKReceipt := sumKReceipt(alws)
	alwDonate := sumDonation(alws)
	return alwKReceipt + alwDonate
}

func validateIncomeExpense(ie request.IncomeExpense) error {
	err := validateAllowance(ie.Allowances)
	if err != nil {
		return err
	}
	err = validateIncome(ie.TotalIncome)
	if err != nil {
		return err
	}

	err = validateWht(ie.TotalIncome, ie.Wht)
	if err != nil {
		return err
	}
	return nil
}

func validateAllowance(alws []request.Allowance) error {
	for _, alw := range alws {
		if alw.Amount < 0 {
			return errors.New("Amount allowance must greater than 0.")
		}
		switch alw.AllowanceType {
		case "donation", "k-receipt":
			continue
		default:
			return errors.New("AllowanceType is 'donation' or 'k-receipt' only")
		}
	}
	return nil
}

func validateWht(income, wht float64) error {
	if wht < 0 || wht > income {
		return errors.New("Wht must be in the range 0 to TotalIncome.")
	}
	return nil
}
func validateIncome(income float64) error {
	if income < 0 {
		return errors.New("TotalIncome must have a starting value of 0.")
	}
	return nil
}

func calculateTaxLevel(income float64, tConst TaxConst) (taxLevel float64) {
	taxInLevel := (income - tConst.Lower) * float64(tConst.TaxRate) / 100
	return float64(taxInLevel)
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
		return c.JSON(http.StatusBadRequest, Err{err.Error()})
	}

	hRecord, _ := reader.GetLine()
	err = validateHeadCSV(hRecord)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{err.Error()})
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
		donate = getMinDonation(donate)

		iNet := calculateIncome(income, donate, personalD)
		netTax, refund := calculateTaxNet(iNet, wht)

		taxInc = resp.TaxWithIncome{TotalIncome: income, Tax: netTax, TaxRefund: refund}

		taxes = append(taxes, taxInc)

	}

	return c.JSON(http.StatusOK, resp.Taxes{Taxes: taxes})
}

func validateHeadCSV(headers []string) error {
	if len(headers) != 3 {
		return errors.New("Header of content is 'totalIncome,wht,donation' only")
	}
	if headers[0] != "totalIncome" || headers[1] != "wht" || headers[2] != "donation" {
		return errors.New("Header of content is 'totalIncome,wht,donation' only")
	}
	return nil
}

func calculateIncome(income, totalAlw, PersonalDed float64) float64 {
	alwTotal := totalAlw

	iNet := income - PersonalDed
	iNet -= alwTotal
	return iNet
}

func sumDonation(alws []request.Allowance) float64 {
	alwTotal := 0.0
	for _, alw := range alws {
		if alw.AllowanceType == "donation" {
			alwTotal += alw.Amount
		}
	}
	alwTotal = getMinDonation(alwTotal)
	return alwTotal
}
func sumKReceipt(alws []request.Allowance) float64 {
	alwTotal := 0.0
	for _, alw := range alws {
		if alw.AllowanceType == "k-receipt" {
			alwTotal += alw.Amount
		}
	}
	alwTotal = math.Min(alwTotal, 50000.0)
	return alwTotal
}

func getMinDonation(donation float64) float64 {
	return math.Min(donation, 100000)
	// return math.Min(donation, 100000000000)
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
				taxInLevel := (tConst.Upper - tConst.Lower) * float64(tConst.TaxRate) / 100
				ttax += taxInLevel

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
