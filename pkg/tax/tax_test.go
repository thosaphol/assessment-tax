package tax

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	req "github.com/thosaphol/assessment-tax/pkg/request"
	resp "github.com/thosaphol/assessment-tax/pkg/response"
)

func TestIncomeExpenseValidation(t *testing.T) {
	tt := []struct {
		name     string
		ie       any
		wantCode int
		wantBody any
	}{
		{
			name: "given amount allowance is less than 0 to calculate tax should return code 400 and message",
			ie: req.IncomeExpense{
				Allowances: []req.Allowance{
					{AllowanceType: "donation", Amount: -1},
				},
			},
			wantCode: http.StatusBadRequest,
			wantBody: Err{Message: "Amount allowance must greater than 0."},
		},
		{
			name: "given amount allowance is 0 to calculate tax should return code 200",
			ie: req.IncomeExpense{
				Allowances: []req.Allowance{
					{AllowanceType: "donation", Amount: 0},
				},
			},
			wantCode: http.StatusOK,
		},

		{
			name: "given income has allowance type is '' to calculate tax should return code 400",
			ie: req.IncomeExpense{
				Allowances: []req.Allowance{
					{AllowanceType: ""},
				},
			},
			wantCode: http.StatusBadRequest,
			wantBody: Err{Message: "AllowanceType is 'donation' only"},
		},
		{
			name: "given income has allowance type isn't 'donation' to calculate tax should return code 400",
			ie: req.IncomeExpense{
				Allowances: []req.Allowance{
					{AllowanceType: "qwerty"},
				},
			},
			wantCode: http.StatusBadRequest,
			wantBody: Err{Message: "AllowanceType is 'donation' only"},
		},
		{
			name: "given income has allowance type is 'donation; to calculate tax should return code 400",
			ie: req.IncomeExpense{
				Allowances: []req.Allowance{
					{AllowanceType: "donation"},
				},
			},
			wantCode: http.StatusOK,
		},

		{
			name: "given income less than 0 to calculate tax should return 400 and message",
			ie: req.IncomeExpense{
				TotalIncome: -1,
				Wht:         0.0,
			},
			wantCode: http.StatusBadRequest,
			wantBody: Err{Message: "TotalIncome must have a starting value of 0."},
		},
		{
			name: "given income than 0 to calculate tax should return 200",
			ie: req.IncomeExpense{
				TotalIncome: 0,
				Wht:         0.0,
			},
			wantCode: http.StatusOK,
		},
		{
			name: "given withholding less than 0 to calculate tax should return code 400 and message",
			ie: req.IncomeExpense{
				TotalIncome: 0,
				Wht:         -1.0,
			},
			wantCode: http.StatusBadRequest,
			wantBody: Err{Message: "Wht must be in the range 0 to TotalIncome."},
		},
		{
			name: "given withholding greater than totalIncome to to calculate tax should return code 400 and message",
			ie: req.IncomeExpense{
				TotalIncome: 100,
				Wht:         1000.0,
			},
			wantCode: http.StatusBadRequest,
			wantBody: Err{Message: "Wht must be in the range 0 to TotalIncome."},
		},
		{
			name: "given withholding,income than 0 to calculate tax should return code 200",
			ie: req.IncomeExpense{
				TotalIncome: 0,
				Wht:         0,
			},
			wantCode: http.StatusOK,
		},
	}

	for _, tCase := range tt {
		t.Run(tCase.name, func(t *testing.T) {
			bytesObj, _ := json.Marshal(tCase.ie)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(bytesObj)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e := echo.New()
			c := e.NewContext(req, rec)
			c.SetPath("/tax/calculations")

			h := New()

			var wantCode = tCase.wantCode
			var wantBody = tCase.wantBody

			h.Calculation(c)
			var gotCode = rec.Code
			var gotBody Err
			gotJson := rec.Body.Bytes()
			if err := json.Unmarshal(gotJson, &gotBody); err != nil {
				t.Errorf("unable to unmarshal json: %v", err)
			}

			if gotCode != wantCode {
				t.Errorf("expected code %v but got code %v", wantCode, gotCode)
			}
			if wantBody != nil && !reflect.DeepEqual(gotBody, wantBody) {
				t.Errorf("expected %v but got %v", wantBody, gotBody)
			}
		})
	}
}

func TestTaxCalculation(t *testing.T) {
	tt := []struct {
		name string
		ie   req.IncomeExpense
		want float64
	}{
		{
			name: "Free tax when income is 0",
			ie: req.IncomeExpense{
				TotalIncome: 0.0,
				Wht:         0.0,
			},
			want: 0,
		},
		{
			name: "Free tax when income is 210,000",
			ie: req.IncomeExpense{
				TotalIncome: 210000,
				Wht:         0.0,
			},
			want: 0,
		},
		{
			name: "tax 0.1 when income is 210,001",
			ie: req.IncomeExpense{
				TotalIncome: 210001,
				Wht:         0.0,
			},
			want: 0.1,
		},
		{
			name: "tax 35,000 when income is 560,000",
			ie: req.IncomeExpense{
				TotalIncome: 560000,
				Wht:         0.0,
			},
			want: 35000,
		},
		{
			name: "tax 35,000.1 when income is 560,001",
			ie: req.IncomeExpense{
				TotalIncome: 560001,
				Wht:         0.0,
			},
			want: 35000 + 0.15,
		},
		{
			name: "tax 110,000 when income is 1,060,000",
			ie: req.IncomeExpense{
				TotalIncome: 1060000,
				Wht:         0.0,
			},
			want: 35000 + 75000,
		},
		{
			name: "tax 110,000.2 when income is 1,060,001",
			ie: req.IncomeExpense{
				TotalIncome: 1060001,
				Wht:         0.0,
			},
			want: 35000 + 75000 + 0.2,
		},
		{
			name: "tax 210,000 when income is 2,060,000",
			ie: req.IncomeExpense{
				TotalIncome: 2060000,
				Wht:         0.0,
			},
			want: 35000 + 75000 + 200000,
		},
		{
			name: "tax 210,000.35 when income is 2,060,001",
			ie: req.IncomeExpense{
				TotalIncome: 2060001,
				Wht:         0.0,
			},
			want: 35000 + 75000 + 200000 + 0.35,
		},
	}

	for _, tCase := range tt {

		t.Run(tCase.name, func(t *testing.T) {

			bytesObj, _ := json.Marshal(tCase.ie)

			req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(string(bytesObj)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e := echo.New()
			c := e.NewContext(req, rec)
			c.SetPath("/tax/calculations")

			h := New()

			var want = resp.Tax{Tax: tCase.want}

			h.Calculation(c)
			var got resp.Tax
			gotJson := rec.Body.Bytes()
			if err := json.Unmarshal(gotJson, &got); err != nil {
				t.Errorf("unable to unmarshal json: %v", err)
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("expected %v but got %v", want, got)
			}

		})
	}
}

func TestTaxCalculationWithWht(t *testing.T) {
	tt := []struct {
		name    string
		ie      req.IncomeExpense
		wantTax any
	}{
		{
			name: "tax 19,000, wiht 0,allowance 200000, when income is 500,000",
			ie: req.IncomeExpense{
				TotalIncome: 500000.0,
				Wht:         0.0,
				Allowances: []req.Allowance{
					{AllowanceType: "donation", Amount: 200000.0},
				},
			},
			wantTax: resp.Tax{Tax: 19000.0},
		},
		{
			name: "tax 22,000, wiht 0,allowance 70,000, when income is 500,000",
			ie: req.IncomeExpense{
				TotalIncome: 500000.0,
				Wht:         0.0,
				Allowances: []req.Allowance{
					{AllowanceType: "donation", Amount: 70000.0},
				},
			},
			wantTax: resp.Tax{Tax: 22000.0},
		},
		{
			name: "tax 22,000, wiht 0,allowance 0, when income is 500,000",
			ie: req.IncomeExpense{
				TotalIncome: 500000.0,
				Wht:         0.0,
				Allowances: []req.Allowance{
					{AllowanceType: "donation", Amount: 0},
				},
			},
			wantTax: resp.Tax{Tax: 29000.0},
		},
		{
			name: "tax 35,000, wiht 0 when income is 560,000",
			ie: req.IncomeExpense{
				TotalIncome: 560000,
				Wht:         0.0,
			},
			wantTax: resp.Tax{Tax: 35000},
		},
		{
			name: "tax 23,000, wiht 12,000 when income is 560,000",
			ie: req.IncomeExpense{
				TotalIncome: 560000,
				Wht:         12000.0,
			},
			wantTax: resp.Tax{Tax: 23000},
		},
		{
			name: "tax 0, wiht 40,000 when income is 560,000",
			ie: req.IncomeExpense{
				TotalIncome: 560000,
				Wht:         40000.0,
			},
			wantTax: resp.TaxWithRefund{Tax: resp.Tax{Tax: 0}, TaxRefund: 5000},
		},
	}

	for _, tCase := range tt {

		t.Run(tCase.name, func(t *testing.T) {

			bytesObj, _ := json.Marshal(tCase.ie)

			req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(string(bytesObj)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e := echo.New()
			c := e.NewContext(req, rec)
			c.SetPath("/tax/calculations")

			h := New()

			var wantTax = tCase.wantTax

			h.Calculation(c)
			var got any
			gotJson := rec.Body.Bytes()
			if reflect.TypeOf(wantTax) == reflect.TypeOf(resp.Tax{}) {
				var taxGot resp.Tax
				if err := json.Unmarshal(gotJson, &taxGot); err != nil {
					t.Errorf("unable to unmarshal json: %v", err)
				}
				got = taxGot
			} else if reflect.TypeOf(wantTax) == reflect.TypeOf(resp.TaxWithRefund{}) {
				var taxGot resp.TaxWithRefund
				if err := json.Unmarshal(gotJson, &taxGot); err != nil {
					t.Errorf("unable to unmarshal json: %v", err)
				}
				got = taxGot
			}

			if !reflect.DeepEqual(got, wantTax) {
				t.Errorf("expected %v but got %v", wantTax, got)
			}

		})
	}
}
