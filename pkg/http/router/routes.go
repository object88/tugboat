package router

import (
	"io"
	"net/http"

	"github.com/object88/tugboat/pkg/http/router/route"
)

func Defaults(subroutes ...[]*route.Route) []*route.Route {
	routes := []*route.Route{
		{
			Path:    "/",
			Handler: DefaultHandleRoot,
			Methods: []string{http.MethodGet},
		},
		{
			Path:    "/liveness",
			Handler: DefaultHandleLiveness,
			Methods: []string{http.MethodGet},
		},
		{
			Path:    "/readiness",
			Handler: DefaultHandleReadiness,
			Methods: []string{http.MethodGet},
		},
	}

	for _, subroute := range subroutes {
		routes = append(routes, subroute...)
	}

	return routes
}

func DefaultHandleRoot(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "shipyard")
}

func DefaultHandleLiveness(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "OK")
}

func DefaultHandleReadiness(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "OK")
}
