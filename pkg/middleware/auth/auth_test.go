package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestBasicAuth(t *testing.T) {

	tt := []struct {
		name     string
		username string
		password string
		wantCode int
	}{
		{
			name:     "Response code 401 when username is incorrect",
			username: "qwwe",
			password: "admin!",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "Response code 401 when password is incorrect",
			username: "adminTax",
			password: "cdf",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "Response code 401 when username and password is incorrect",
			username: "cdcvd",
			password: "cdf",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "Response code 200 when username and password is correct",
			username: "adminTax",
			password: "admin!",
			wantCode: http.StatusOK,
		},
	}

	user := "adminTax"
	pass := "admin!"

	for _, tCase := range tt {
		t.Run(tCase.name, func(t *testing.T) {

			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			rec := httptest.NewRecorder()
			req.SetBasicAuth(tCase.username, tCase.password)
			e := echo.New()
			c := e.NewContext(req, rec)
			c.SetPath("/admin")

			e.Use(NewBasicAuth(user, pass))
			e.GET("/admin", func(c echo.Context) error {
				return c.JSON(http.StatusOK, "You are authorized!")
			})

			var wantCode = tCase.wantCode

			e.ServeHTTP(rec, req)
			var gotCode = rec.Code

			if gotCode != wantCode {
				t.Errorf("expected code %v but got code %v", wantCode, gotCode)
			}
		})
	}
}
