package deduction

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/thosaphol/assessment-tax/pkg/request"
	req "github.com/thosaphol/assessment-tax/pkg/request"
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
func (stubStore StubStore) PersonalDeduction() (float64, error) {
	return stubStore.deduction.Personal, stubStore.err
}

func (stubStore StubStore) SetKReceiptDeduction(amount float64) error {
	return stubStore.err
}

func (stubStore StubStore) KReceiptDeduction() (float64, error) {
	return stubStore.deduction.MaxKReceipt, stubStore.err
}

func TestPersonalDeductionValidation(t *testing.T) {
	tt := []struct {
		name     string
		d        any
		wantCode int
		wantBody any
	}{
		{
			name:     "given incorrect structure should return code 400 and message",
			d:        resp.Tax{},
			wantCode: http.StatusBadRequest,
			wantBody: resp.Err{Message: "Json structure invalid"},
		},
		{
			name:     "given amount deduction less than 0 should return code 400 and message",
			d:        req.PersonalDeduction{Amount: -1},
			wantCode: http.StatusBadRequest,
			wantBody: resp.Err{Message: "Amount: Invalid amount is required 10,000.0 to 100,000.0"},
		},
		{
			name:     "given amount deduction greater than 100,000.0 should return code 400 and message",
			d:        req.PersonalDeduction{Amount: 100001.0},
			wantCode: http.StatusBadRequest,
			wantBody: resp.Err{Message: "Amount: Invalid amount is required 10,000.0 to 100,000.0"},
		},
		{
			name:     "given amount deduction in range 10,000.0 to 100,000.0 should return code 200 and response",
			d:        req.PersonalDeduction{Amount: 10200.0},
			wantCode: http.StatusOK,
		},
	}

	stubStore := StubStore{
		deduction: Deduction{},
		err:       nil,
	}

	for _, tCase := range tt {
		t.Run(tCase.name, func(t *testing.T) {
			bytesObj, _ := json.Marshal(tCase.d)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(bytesObj)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e := echo.New()
			c := e.NewContext(req, rec)
			c.SetPath("/admin/deductions/personal")

			h := New(stubStore)

			var wantCode = tCase.wantCode
			var wantBody = tCase.wantBody

			h.SetDeductionPersonal(c)
			var gotCode = rec.Code
			var gotBody resp.Err
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

func TestKReceiptDeductionValidation(t *testing.T) {
	tt := []struct {
		name     string
		d        any
		wantCode int
		wantBody any
	}{
		{
			name:     "given incorrect structure should return code 400 and message",
			d:        resp.Tax{},
			wantCode: http.StatusBadRequest,
			wantBody: resp.Err{Message: "Json structure invalid"},
		},
		{
			name:     "given amount deduction less than 0 should return code 400 and message",
			d:        req.KReceiptDeduction{Amount: -1},
			wantCode: http.StatusBadRequest,
			wantBody: resp.Err{Message: "Amount: Invalid amount is required 0.0 to 100,000.0"},
		},
		{
			name:     "given amount deduction greater than 100,000.0 should return code 400 and message",
			d:        req.PersonalDeduction{Amount: 100001.0},
			wantCode: http.StatusBadRequest,
			wantBody: resp.Err{Message: "Amount: Invalid amount is required 0.0 to 100,000.0"},
		},
		{
			name:     "given amount deduction is 0.0 should return code 200",
			d:        req.PersonalDeduction{Amount: 0.0},
			wantCode: http.StatusOK,
		},
		{
			name:     "given amount deduction is 100,000.0 should return code 200",
			d:        req.PersonalDeduction{Amount: 100000.0},
			wantCode: http.StatusOK,
		},
	}

	stubStore := StubStore{
		deduction: Deduction{},
		err:       nil,
	}

	for _, tCase := range tt {
		t.Run(tCase.name, func(t *testing.T) {
			bytesObj, _ := json.Marshal(tCase.d)

			req := httptest.NewRequest(http.MethodPost, "/admin/deductions/k-receipt", strings.NewReader(string(bytesObj)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e := echo.New()
			c := e.NewContext(req, rec)
			c.SetPath("/admin/deductions/k-receipt")

			h := New(stubStore)

			var wantCode = tCase.wantCode
			var wantBody = tCase.wantBody

			h.SetDeductionKReceipt(c)
			var gotCode = rec.Code
			var gotBody resp.Err
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

func TestDeductionError(t *testing.T) {
	stubStore := StubStore{
		deduction: Deduction{},
		err:       errors.New("database error"),
	}

	t.Run("given correct deduction should return code 500 and message", func(t *testing.T) {
		body := req.PersonalDeduction{Amount: 10200.0}

		bytesObj, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/admin/deductions/personal", strings.NewReader(string(bytesObj)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e := echo.New()
		c := e.NewContext(req, rec)
		c.SetPath("/admin/deductions/personal")

		h := New(stubStore)

		var wantCode = http.StatusInternalServerError
		var wantBody = resp.Err{Message: "Found Internal Server Error"}

		h.SetDeductionPersonal(c)
		var gotCode = rec.Code
		var gotBody resp.Err
		gotJson := rec.Body.Bytes()
		if err := json.Unmarshal(gotJson, &gotBody); err != nil {
			t.Errorf("unable to unmarshal json: %v", err)
		}

		if gotCode != wantCode {
			t.Errorf("expected code %v but got code %v", wantCode, gotCode)
		}
		if !reflect.DeepEqual(gotBody, wantBody) {
			t.Errorf("expected %v but got %v", wantBody, gotBody)
		}
	})

	t.Run("given correct k-receipt deduction should return code 500 and message", func(t *testing.T) {
		body := req.PersonalDeduction{Amount: 10200.0}

		bytesObj, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/admin/deductions/k-receipt", strings.NewReader(string(bytesObj)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e := echo.New()
		c := e.NewContext(req, rec)
		c.SetPath("/admin/deductions/k-receipt")

		h := New(stubStore)

		var wantCode = http.StatusInternalServerError
		var wantBody = resp.Err{Message: "Found Internal Server Error"}

		h.SetDeductionKReceipt(c)
		var gotCode = rec.Code
		var gotBody resp.Err
		gotJson := rec.Body.Bytes()
		if err := json.Unmarshal(gotJson, &gotBody); err != nil {
			t.Errorf("unable to unmarshal json: %v", err)
		}

		if gotCode != wantCode {
			t.Errorf("expected code %v but got code %v", wantCode, gotCode)
		}
		if !reflect.DeepEqual(gotBody, wantBody) {
			t.Errorf("expected %v but got %v", wantBody, gotBody)
		}
	})
}

func TestPersonalDeduction(t *testing.T) {
	tt := []struct {
		name     string
		d        request.PersonalDeduction
		wantCode int
		wantBody any
	}{
		{
			name:     "personal deduction is 70000 when amount is 70000",
			d:        request.PersonalDeduction{Amount: 70000},
			wantCode: http.StatusOK,
			wantBody: response.PersonalDeduction{PersonalDeduction: 70000},
		},
		{
			name:     "personal deduction is 50000 when amount is 50000",
			d:        request.PersonalDeduction{Amount: 50000},
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

func TestKReceiptDeduction(t *testing.T) {
	tt := []struct {
		name     string
		d        request.KReceiptDeduction
		wantCode int
		wantBody any
	}{
		{
			name:     "k-receipt deduction is 70000 when amount is 70000",
			d:        request.KReceiptDeduction{Amount: 70000},
			wantCode: http.StatusOK,
			wantBody: response.KReceiptDeduction{KReceipt: 70000},
		},
		{
			name:     "k-receipt deduction is 50000 when amount is 50000",
			d:        request.KReceiptDeduction{Amount: 50000},
			wantCode: http.StatusOK,
			wantBody: response.KReceiptDeduction{KReceipt: 50000},
		},
	}

	stubstore := StubStore{
		deduction: Deduction{},
		err:       nil,
	}

	for _, tCase := range tt {
		t.Run(tCase.name, func(t *testing.T) {
			bytesObj, _ := json.Marshal(tCase.d)

			req := httptest.NewRequest(http.MethodGet, "/admin/deductions/k-receip", strings.NewReader(string(bytesObj)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e := echo.New()
			c := e.NewContext(req, rec)
			c.SetPath("/admin/deductions/k-receipt")

			h := New(stubstore)

			var wantBody = tCase.wantBody
			var wantCode = tCase.wantCode

			h.SetDeductionKReceipt(c)
			var gotCode = rec.Code
			var gotBody resp.KReceiptDeduction

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
