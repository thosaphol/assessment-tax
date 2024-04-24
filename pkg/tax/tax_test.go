package tax

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/thosaphol/assessment-tax/pkg/request"
	"github.com/thosaphol/assessment-tax/pkg/response"
)

func TestTaxCalculation(t *testing.T) {
	tt := []struct {
		name string
		ie   request.IncomeExpense
		want float64
	}{
		{
			name: "Free tax when income is 0",
			ie: request.IncomeExpense{
				TotalIncome: 0.0,
				Wht:         0.0,
			},
			want: 0,
		},
		{
			name: "Free tax when income is 210,000",
			ie: request.IncomeExpense{
				TotalIncome: 210000,
				Wht:         0.0,
			},
			want: 0,
		},
		{
			name: "tax 0.1 when income is 210,001",
			ie: request.IncomeExpense{
				TotalIncome: 210001,
				Wht:         0.0,
			},
			want: 0.1,
		},
		{
			name: "tax 35,000 when income is 560,000",
			ie: request.IncomeExpense{
				TotalIncome: 560000,
				Wht:         0.0,
			},
			want: 35000,
		},
		{
			name: "tax 35,000.1 when income is 560,001",
			ie: request.IncomeExpense{
				TotalIncome: 560001,
				Wht:         0.0,
			},
			want: 35000 + 0.15,
		},
		{
			name: "tax 110,000 when income is 1,060,000",
			ie: request.IncomeExpense{
				TotalIncome: 1060000,
				Wht:         0.0,
			},
			want: 35000 + 75000,
		},
		{
			name: "tax 110,000.2 when income is 1,060,001",
			ie: request.IncomeExpense{
				TotalIncome: 1060001,
				Wht:         0.0,
			},
			want: 35000 + 75000 + 0.2,
		},
		{
			name: "tax 210,000 when income is 2,060,000",
			ie: request.IncomeExpense{
				TotalIncome: 2060000,
				Wht:         0.0,
			},
			want: 35000 + 75000 + 200000,
		},
		{
			name: "tax 210,000.35 when income is 2,060,001",
			ie: request.IncomeExpense{
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

			var want = response.Tax{Tax: tCase.want}

			h.Calculation(c)
			var got response.Tax
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
