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
	"github.com/thosaphol/assessment-tax/pkg/tax"
)

var (
	ENV_PORT = "PORT"
)

func main() {

	var port = "8080"
	_, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal("PORT Variable is not an integer.")
		return
	}

	h := tax.New()

	e := echo.New()
	e.POST("tax/calculations", h.Calculation)

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
