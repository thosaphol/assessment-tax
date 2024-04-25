package deduction

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
	resp "github.com/thosaphol/assessment-tax/pkg/response"
)

type StubStore struct {
	deduction Deduction
	err       error
}

// Wallets implements Storer.
func (stubStore StubStore) SetPersonalDeduction(amount float64) error {
	return stubStore.err
}

func TestPersonalDeduction(t *testing.T) {
	tt := []struct {
		name     string
		d        request.Deduction
		wantCode int
		wantBody any
	}{
		{
			name:     "personal deduction is 70000 when amount is 70000",
			d:        request.Deduction{Amount: 70000},
			wantCode: http.StatusOK,
			wantBody: response.PersonalDeduction{PersonalDeduction: 70000},
		},
		{
			name:     "personal deduction is 50000 when amount is 50000",
			d:        request.Deduction{Amount: 50000},
			wantCode: http.StatusOK,
			wantBody: response.PersonalDeduction{PersonalDeduction: 50000},
		},
	}

	stubstore := StubStore{
		deduction: Deduction{},
		err:       nil,
	}

	for _, tCase := range tt {
		t.Run(tCase.name, func(t *testing.T) {
			bytesObj, _ := json.Marshal(tCase.d)

			req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(string(bytesObj)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e := echo.New()
			c := e.NewContext(req, rec)
			c.SetPath("/admin/deductions/personal")

			h := New(stubstore)

			var wantBody = tCase.wantBody
			var wantCode = tCase.wantCode

			h.SetDeductionPersonal(c)
			var gotCode = rec.Code
			var gotBody resp.PersonalDeduction

			gotJson := rec.Body.Bytes()
			if err := json.Unmarshal(gotJson, &gotBody); err != nil {
				t.Errorf("unable to unmarshal json: %v", err)
			}

			if wantCode != gotCode {
				t.Errorf("expected code %v but got code %v", wantCode, gotCode)
			}
			if !reflect.DeepEqual(gotBody, wantBody) {
				t.Errorf("expected %v but got %v", wantBody, gotBody)
			}

		})
	}

}
