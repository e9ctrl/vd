package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/e9ctrl/vd/log"
	"github.com/go-chi/chi/v5"
	"github.com/jwalton/gchalk"
)

type Device interface {
	GetParameter(param string) (any, error)
	SetParameter(param string, val any) error
	GetGlobalDelay(typ string) (time.Duration, error)
	SetGlobalDelay(typ string, val string) error
	GetParamDelay(typ string, param string) (time.Duration, error)
	SetParamDelay(typ string, param string, val string) error
	GetMismatch() []byte
	SetMismatch(string) error
	Trigger(param string) error
}

type api struct {
	d Device
	http.Server
}

func NewHTTP(d Device, addr string) *api {

	return &api{
		d: d,
		Server: http.Server{
			Addr: addr,
		},
	}
}
func (a *api) getMismatch(w http.ResponseWriter, r *http.Request) {
	value := a.d.GetMismatch()
	log.API("get mismatch")
	w.Header().Set("Content-Type", "text/plain")
	w.Write(value)
}

func (a *api) setMismatch(w http.ResponseWriter, r *http.Request) {
	value := chi.URLParam(r, "value")

	err := a.d.SetMismatch(value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.API("set mismatch to", value)
	w.Write([]byte("Mismatch set successfully"))
}
func (a *api) getGlobalDelay(w http.ResponseWriter, r *http.Request) {
	typ := chi.URLParam(r, "type")

	del, err := a.d.GetGlobalDelay(typ)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}
	log.API("get delay", typ)
	// Return the value as plain text
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(del.String()))
}

func (a *api) setGlobalDelay(w http.ResponseWriter, r *http.Request) {
	typ := chi.URLParam(r, "type")
	value := chi.URLParam(r, "value")

	err := a.d.SetGlobalDelay(typ, value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.API("set delay", typ, "to", value)
	w.Write([]byte("Delay set successfully"))
}

func (a *api) getParameter(w http.ResponseWriter, r *http.Request) {
	param := chi.URLParam(r, "param")

	value, err := a.d.GetParameter(param)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}
	log.API("get", param)
	// Return the value as plain text
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(fmt.Sprintf("%v", value)))
}

func (a *api) setParameter(w http.ResponseWriter, r *http.Request) {
	param := chi.URLParam(r, "param")
	value := chi.URLParam(r, "value")

	err := a.d.SetParameter(param, value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.API("set", param, "to", value)
	w.Write([]byte("Parameter set successfully"))
}

func (a *api) getParamDelay(w http.ResponseWriter, r *http.Request) {
	typ := chi.URLParam(r, "type")
	param := chi.URLParam(r, "param")

	del, err := a.d.GetParamDelay(param, typ)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	log.API("get delay", typ, "of", param)
	// Return the value as plain text
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(del.String()))
}

func (a *api) setParamDelay(w http.ResponseWriter, r *http.Request) {
	typ := chi.URLParam(r, "type")
	param := chi.URLParam(r, "param")
	value := chi.URLParam(r, "value")

	err := a.d.SetParamDelay(param, typ, value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.API("set delay", typ, "of", param, "to", value)
	w.Write([]byte("Delay set successfully"))
}

func (a *api) trigger(w http.ResponseWriter, r *http.Request) {
	param := chi.URLParam(r, "param")

	err := a.d.Trigger(param)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.API("triggered parameter", param)
	w.Write([]byte("Parameter triggered successfully"))
}

func (a *api) Start() {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/{param}", a.getParameter)
		r.Post("/{param}/{value}", a.setParameter)
		r.Get("/delay/{type}", a.getGlobalDelay)
		r.Post("/delay/{type}/{value}", a.setGlobalDelay)
		r.Get("/delay/{type}/{param}", a.getParamDelay)
		r.Post("/delay/{type}/{param}/{value}", a.setParamDelay)
		r.Get("/mismatch", a.getMismatch)
		r.Post("/mismatch/{value}", a.setMismatch)
		r.Post("/trigger/{param}", a.trigger)
	})

	fmt.Println("HTTP API listening on ", gchalk.BrightMagenta("http://"+a.Addr))

	a.Handler = r
	a.ListenAndServe()
}
