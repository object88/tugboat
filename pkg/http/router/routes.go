package router

import (
	"io"
	"net/http"

	"github.com/object88/tugboat/pkg/http/probes"
	"github.com/object88/tugboat/pkg/http/router/route"
)

func Defaults(p *probes.Probe, subroutes ...[]*route.Route) []*route.Route {
	routes := []*route.Route{
		{
			Path:    "/",
			Handler: DefaultHandleRoot,
			Methods: []string{http.MethodGet},
		},
		{
			Path:    "/liveness",
			Handler: DefaultHandleLiveness(p),
			Methods: []string{http.MethodGet},
		},
		{
			Path:    "/readiness",
			Handler: DefaultHandleReadiness(p),
			Methods: []string{http.MethodGet},
		},
	}

	for _, subroute := range subroutes {
		routes = append(routes, subroute...)
	}

	return routes
}

func DefaultHandleRoot(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "tugboat")
}

func DefaultHandleLiveness(p *probes.Probe) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if p.IsLive() {
			io.WriteString(w, "OK")
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func DefaultHandleReadiness(p *probes.Probe) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if p.IsReady() {
			io.WriteString(w, "OK")
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}
