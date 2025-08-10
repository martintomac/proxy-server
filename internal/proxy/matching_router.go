package proxy

import "net/http"

type MatchingRouter struct {
	routes []matchingRoute
}

type matchingRoute struct {
	predicate MatchingPredicate
	handler   Handler
}

func NewMatchingRouter() *MatchingRouter {
	return &MatchingRouter{}
}

func (mr *MatchingRouter) AddRoute(predicate MatchingPredicate, handler Handler) {
	mr.routes = append(mr.routes, matchingRoute{
		predicate: predicate,
		handler:   handler,
	})
}

func (mr *MatchingRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range mr.routes {
		if route.predicate.match(r) {
			route.handler.ServeHTTP(w, r)
			return
		}
	}
}
