package router

import (
	"fmt"
	"strings"

	"github.com/gorilla/mux"
	"github.com/object88/tugboat/pkg/http/router/route"
	"github.com/sirupsen/logrus"
)

type Router struct {
	m      *mux.Router
	logger *logrus.Logger
}

// New creates a new Router
func New(logger *logrus.Logger) *Router {
	rtr := &Router{
		m:      mux.NewRouter(),
		logger: logger,
	}
	return rtr
}

func (rtr *Router) Route(routes []*route.Route) (*mux.Router, error) {
	if err := rtr.configureRoutes(rtr.m, routes); err != nil {
		return nil, err
	}

	rtr.reportRoutes()

	return rtr.m, nil
}

func (rtr *Router) reportRoutes() {
	rtr.m.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		rtr.logger.Infof("Name: %s\n", route.GetName())

		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			rtr.logger.Infof("Route: %s\n", pathTemplate)
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			rtr.logger.Infof("Path regexp: %s\n", pathRegexp)
		}
		queriesTemplates, err := route.GetQueriesTemplates()
		if err == nil {
			rtr.logger.Infof("Queries templates: %s\n", strings.Join(queriesTemplates, ","))
		}
		queriesRegexps, err := route.GetQueriesRegexp()
		if err == nil {
			rtr.logger.Infof("Queries regexps: %s\n", strings.Join(queriesRegexps, ","))
		}
		if route.GetHandler() != nil {
			rtr.logger.Infof("Has router func\n")
		}
		rtr.logger.Infof("\n")
		return nil
	})
}

func (rtr *Router) configureRoutes(base *mux.Router, routes []*route.Route) error {
	for _, r := range routes {
		if r.Handler != nil {
			s := base.NewRoute().Subrouter()

			if len(r.Middleware) != 0 {
				s.Use(r.Middleware...)
			}

			rt := s.Path(r.Path)
			rt = rt.Handler(r.Handler)

			if len(r.Methods) != 0 {
				rt = rt.Methods(r.Methods...)
			}
			if len(r.Queries) != 0 {
				qs := make([]string, len(r.Queries)*2)
				for k, v := range r.Queries {
					qs[k*2] = v
					qs[k*2+1] = fmt.Sprintf("{%s}", v)
				}
				rt = rt.Queries(qs...)
			}
			err := rt.GetError()
			if err != nil {
				return fmt.Errorf("failed to create route for path '%s': %w", r.Path, err)
			}
		}

		if len(r.Subroutes) != 0 {
			sub := base.PathPrefix(r.Path).Subrouter()

			if len(r.Middleware) != 0 {
				for _, mfunc := range r.Middleware {
					sub.Use(mfunc)
				}
			}

			rtr.configureRoutes(sub, r.Subroutes)
		}
	}
	return nil
}
