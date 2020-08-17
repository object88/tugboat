package route

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Route struct {
	Path       string
	Handler    http.HandlerFunc
	Methods    []string
	Queries    []string
	Middleware []mux.MiddlewareFunc
	Subroutes  []*Route
}
