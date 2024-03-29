package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/e9ctrl/vd/log"
	"github.com/go-chi/chi/v5"
	"github.com/jwalton/gchalk"
)

// Interface provides methods to control parameters and commands settings via HTTP server.
type Device interface {
	GetParameter(param string) (any, error)
	SetParameter(param string, val any) error
	GetCommandDelay(commandName string) (time.Duration, error)
	SetCommandDelay(commandName string, val string) error
	GetMismatch() []byte
	SetMismatch(mismatch string) error
	Trigger(param string) error
}

// Struct that keeps Device interface.
type Api struct {
	d Device
}

// Create new instance of http server that fullfils Device interface.
// Server exposes REST API to control parameters and commands settings.
func NewHttpApiServer(d Device) *Api {
	return &Api{d: d}
}

// Start HTTP server
func (a *Api) Serve(ctx context.Context, addr string) error {
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

func (a *Api) routes() http.Handler {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/{param}", a.getParameter)
		r.Post("/{param}/{value}", a.setParameter)
		r.Get("/delay/{command}", a.getCommandDelay)
		r.Post("/delay/{command}/{value}", a.setCommandDelay)
		r.Get("/mismatch", a.getMismatch)
		r.Post("/mismatch/{value}", a.setMismatch)
		r.Post("/trigger/{param}", a.trigger)
	})

	return r
}
func (a *Api) getMismatch(w http.ResponseWriter, r *http.Request) {
	value := a.d.GetMismatch()
	log.API("get mismatch")
	w.Header().Set("Content-Type", "text/plain")
	w.Write(value)
}

func (a *Api) setMismatch(w http.ResponseWriter, r *http.Request) {
	value := chi.URLParam(r, "value")

	err := a.d.SetMismatch(value)
	if err != nil {
		errorHandler(w, err)
		return
	}

	log.API("set mismatch to", value)
	w.Write([]byte("Mismatch set successfully"))
}

func (a *Api) getParameter(w http.ResponseWriter, r *http.Request) {
	param := chi.URLParam(r, "param")

	value, err := a.d.GetParameter(param)
	if err != nil {
		errorHandler(w, err)
		return
	}
	log.API("get", param)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(fmt.Sprintf("%v", value)))
}

func (a *Api) setParameter(w http.ResponseWriter, r *http.Request) {
	param := chi.URLParam(r, "param")
	value := chi.URLParam(r, "value")

	err := a.d.SetParameter(param, value)
	if err != nil {
		errorHandler(w, err)
		return
	}
	log.API("set", param, "to", value)
	w.Write([]byte("Parameter set successfully"))
}

func (a *Api) getCommandDelay(w http.ResponseWriter, r *http.Request) {
	commandName := chi.URLParam(r, "command")

	del, err := a.d.GetCommandDelay(commandName)
	if err != nil {
		errorHandler(w, err)
		return
	}

	log.API("get delay of", commandName)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(del.String()))
}

func (a *Api) setCommandDelay(w http.ResponseWriter, r *http.Request) {
	commandName := chi.URLParam(r, "command")
	value := chi.URLParam(r, "value")

	err := a.d.SetCommandDelay(commandName, value)
	if err != nil {
		errorHandler(w, err)
		return
	}

	log.API("set delay of", commandName, "to", value)
	w.Write([]byte("Delay set successfully"))
}

func (a *Api) trigger(w http.ResponseWriter, r *http.Request) {
	param := chi.URLParam(r, "param")

	err := a.d.Trigger(param)
	if err != nil {
		errorHandler(w, err)
		return
	}

	log.API("triggered parameter", param)
	w.Write([]byte("Parameter triggered successfully"))
}

func errorHandler(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "Error: %s", err)
}
