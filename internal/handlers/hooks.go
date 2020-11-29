package handlers

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func InterruptHandler(errc chan<- error, cancelFunc context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	terminateError := fmt.Errorf("%s", <-c)

	cancelFunc()
	// Place shutdown handling you want here

	errc <- terminateError
}
