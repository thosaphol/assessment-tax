package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/thosaphol/assessment-tax/pkg/deduction"
	"github.com/thosaphol/assessment-tax/pkg/middleware/auth"
	"github.com/thosaphol/assessment-tax/pkg/repo/postgres"
	"github.com/thosaphol/assessment-tax/pkg/tax"
)

var (
	ENV_PORT           = "PORT"
	ENV_DATABASE_URL   = "DATABASE_URL"
	ENV_ADMIN_USERNAME = "ADMIN_USERNAME"
	ENV_ADMIN_PASSWORD = "ADMIN_PASSWORD"
)

func main() {

	var port = os.Getenv(ENV_PORT)
	var connString = os.Getenv(ENV_DATABASE_URL)
	var user = os.Getenv(ENV_ADMIN_USERNAME)
	var pass = os.Getenv(ENV_ADMIN_PASSWORD)
	_, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal("PORT Variable is not an integer.")
		return
	}

	p, err := postgres.New(connString)
	if err != nil {
		log.Fatal(err)
		return
	}

	hd := deduction.New(p)
	h := tax.New(p)

	e := echo.New()
	e.POST("/tax/calculations", h.Calculation)

	g := e.Group("/admin")
	g.Use(auth.NewBasicAuth(user, pass))
	g.POST("/admin/deductions/personal", hd.SetDeductionPersonal)

	//
	// graceful shutdown
	//

	// start server in go routine
	go func() {
		if err := e.Start(fmt.Sprintf(":%s", port)); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal(err)
		}
	}()

	//receive os interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-shutdown

	// create timeout context
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// shutdown server after interrupt
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	fmt.Print("shutting down the server")
	// e.Logger.Fatal(e.Start(":1323"))

}
