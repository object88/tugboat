package v1

import (
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/object88/tugboat/internal/slack"
	"github.com/object88/tugboat/pkg/http/router/route"
	"github.com/object88/tugboat/pkg/logging"
)

func Defaults(logger logr.Logger, bot *slack.Bot) []*route.Route {
	return []*route.Route{
		{
			Path:       "/v1/api",
			Middleware: []mux.MiddlewareFunc{configureLoggingMiddleware(logger)},
			Subroutes: []*route.Route{
				{
					Path:    "/commands",
					Handler: configureHandleCommand(bot),
					Methods: []string{http.MethodPost},
				},
				{
					Path:    "/events",
					Handler: configureHandleEvents(bot),
					Methods: []string{http.MethodPost},
				},
				{
					Path:    "/interactive",
					Handler: configureHandleInteractive(bot),
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

func configureHandleCommand(bot *slack.Bot) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bot.ProcessSlashCommand(w, r)
	}
}

func configureHandleEvents(bot *slack.Bot) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bot.ProcessEventCommand(w, r)
	}
}

func configureHandleInteractive(bot *slack.Bot) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bot.ProcessInteractiveCommand(w, r)
	}
}

type LogContextHandler struct {
	logger logr.Logger
	next   http.Handler
}

func (lch *LogContextHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	creq := req.WithContext(logging.ToContext(req.Context(), lch.logger))
	lch.next.ServeHTTP(w, creq)
}
