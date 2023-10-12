package api

import (
	"fmt"
	"net/http"

	"github.com/e9ctrl/vd/log"
	"github.com/go-chi/chi/v5"
	"github.com/jwalton/gchalk"
)

type Device interface {
	GetParameter(string) (any, error)
	SetParameter(string, any) error
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

func (a *api) Start() {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/{param}", a.getParameter)
		r.Post("/{param}/{value}", a.setParameter)
	})

	fmt.Println("HTTP API listening on ", gchalk.BrightMagenta("http://"+a.Addr))

	a.Handler = r
	a.ListenAndServe()
}
