package proxy

import (
	"net/http"
)

type Router interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type PathRouter struct {
	Routes map[string]Handler
	mux    *http.ServeMux
}

func NewPathRouter() *PathRouter {
	return &PathRouter{
		Routes: make(map[string]Handler),
		mux:    http.NewServeMux(),
	}
}

func (pr *PathRouter) AddRoute(pattern string, handler Handler) {
	pr.Routes[pattern] = handler
	pr.mux.HandleFunc(pattern, handler.ServeHTTP)
}

func (pr *PathRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pr.mux.ServeHTTP(w, r)
}
