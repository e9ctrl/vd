package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jwalton/gchalk"
)

// Start HTTP server
func (a *api) Serve(ctx context.Context, addr string) error {
	srv := &http.Server{
		Addr:         addr,
		Handler:      a.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownError := make(chan error)
	go func() {
		<-ctx.Done()

		ctxChild, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		err := srv.Shutdown(ctxChild)
		if err != nil {
			shutdownError <- err
		}

		shutdownError <- nil
	}()

	fmt.Println("HTTP API listening on ", gchalk.BrightMagenta("http://"+srv.Addr))

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	fmt.Println("server HTTP stopped")
	return nil
}
