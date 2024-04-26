package auth

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewBasicAuth(user, pass string) echo.MiddlewareFunc {
	return middleware.BasicAuth(func(u, p string, ctx echo.Context) (bool, error) {
		if u == user && p == pass {
			return true, nil
		}
		return false, nil
	})

}
