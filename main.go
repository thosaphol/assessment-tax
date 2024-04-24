package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Go Bootcamp!")
	})

	//
	// graceful shutdown
	//

	// start server in go routine
	go func() {
		if err := e.Start(":1323"); err != nil && err != http.ErrServerClosed {
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
