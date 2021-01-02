package v1

import (
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/validator"
	"github.com/object88/tugboat/pkg/http/router/route"
	"github.com/object88/tugboat/pkg/logging"
)

func Defaults(logger logr.Logger, m *validator.M, v *validator.V, v2 *validator.V2) []*route.Route {
	return []*route.Route{
		{
			Path:       "/v1/api",
			Middleware: []mux.MiddlewareFunc{configureLoggingMiddleware(logger)},
			Subroutes: []*route.Route{
				{
					Path:    "/mutate",
					Handler: configureMutatingAdmission(m),
					Methods: []string{http.MethodPost},
				},
				{
					Path:    "/validate",
					Handler: configureValidatingAdmission(v),
					Methods: []string{http.MethodPost},
				},
				{
					Path:    "/validate-helm-secret",
					Handler: configureValidatingHelmSecretAdmission(v2),
					Methods: []string{http.MethodPost},
				},
			},
		},
	}
}

func configureLoggingMiddleware(logger logr.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		lch := LogContextHandler{
			logger: logger,
			next:   next,
		}
		return handlers.LoggingHandler((&logging.Writer{Log: logger}).Out(), &lch)
	}
}

func configureMutatingAdmission(m *validator.M) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.ProcessAdmission(w, r)
	}
}

func configureValidatingAdmission(v *validator.V) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v.ProcessAdmission(w, r)
	}
}

func configureValidatingHelmSecretAdmission(v *validator.V2) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v.ProcessAdmission(w, r)
	}
}

type LogContextHandler struct {
	logger logr.Logger
	next   http.Handler
}

func (lch *LogContextHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	creq := req.WithContext(logr.NewContext(req.Context(), lch.logger))
	lch.next.ServeHTTP(w, creq)
}
