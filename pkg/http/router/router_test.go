package router

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/zapr"
	"github.com/gorilla/mux"
	"github.com/object88/tugboat/pkg/http/router/route"
	"go.uber.org/zap"
)

func Test_Router_Routes(t *testing.T) {
	zapLog, _ := zap.NewDevelopment()
	logger := zapr.NewLogger(zapLog)
	rtr := New(logger)

	defaultRoute := func(_ *Router) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			logger.Info("Unhandled route", "URL", r.URL)
			w.WriteHeader(404)
		}
	}

	routes := []*route.Route{
		{
			Path:    "/",
			Methods: []string{http.MethodGet},
			Handler: func(w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, "OK")
			},
		},
	}
	mux, err := rtr.Route(defaultRoute, routes)
	if err != nil {
		t.Fatal("failed to configure routes")
	}

	ts := httptest.NewServer(mux)
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Errorf("failed to perform get")
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("failed to read body")
	}
	if string(body) != "OK" {
		t.Errorf("incorrect body")
	}
}

func Test_Router_Subroute(t *testing.T) {
	tcs := []struct {
		name   string
		routes []*route.Route
		url    string
	}{
		{
			name: "two segments",
			routes: []*route.Route{
				{
					Path: "/",
					Subroutes: []*route.Route{
						{
							Path: "/a",
							Subroutes: []*route.Route{
								{
									Path:    "/b",
									Methods: []string{http.MethodGet},
									Handler: func(w http.ResponseWriter, r *http.Request) {
										io.WriteString(w, "OK")
									},
								},
							},
						},
					},
				},
			},
			url: "/a/b",
		},
		{
			name: "three segments",
			routes: []*route.Route{
				{
					Path: "/",
					Subroutes: []*route.Route{
						{
							Path: "/a",
							Subroutes: []*route.Route{
								{
									Path: "/b",
									Subroutes: []*route.Route{
										{
											Path:    "/c",
											Methods: []string{http.MethodGet},
											Handler: func(w http.ResponseWriter, r *http.Request) {
												io.WriteString(w, "OK")
											},
										},
									},
								},
							},
						},
					},
				},
			},
			url: "/a/b/c",
		},
		{
			name: "multi-segment",
			routes: []*route.Route{
				{
					Path: "/a/b",
					Subroutes: []*route.Route{
						{
							Path:    "/c",
							Methods: []string{http.MethodGet},
							Handler: func(w http.ResponseWriter, r *http.Request) {
								io.WriteString(w, "OK")
							},
						},
					},
				},
			},
			url: "/a/b/c",
		},
		{
			name: "multi-segment with queries",
			routes: []*route.Route{
				{
					Path: "/a/b",
					Subroutes: []*route.Route{
						{
							Path:    "/c",
							Methods: []string{http.MethodGet},
							Queries: []string{"d", "e"},
							Handler: func(w http.ResponseWriter, r *http.Request) {
								io.WriteString(w, "OK")
							},
						},
					},
				},
			},
			url: "/a/b/c?d=foo&e=bar",
		},
	}

	zapLog, _ := zap.NewDevelopment()
	logger := zapr.NewLogger(zapLog)

	defaultRoute := func(_ *Router) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			logger.Info("Unhandled route", "URL", r.URL)
			w.WriteHeader(404)
		}
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			rtr := New(logger)
			mux, err := rtr.Route(defaultRoute, tc.routes)
			if err != nil {
				t.Fatal("failed to configure routes")
			}

			ts := httptest.NewServer(mux)
			defer ts.Close()

			url := ts.URL + tc.url
			t.Logf("Have URL '%s'", url)

			res, status := get(t, url)
			if res != "OK" {
				t.Errorf("failed to get OK: got '%s'", res)
			}
			if status != http.StatusOK {
				t.Errorf("failed to get OK status: got '%d'", status)
			}
		})
	}
}

func Test_Router_SubrouteWithMiddleware(t *testing.T) {
	tcs := []struct {
		name   string
		routes []*route.Route
		url    string
	}{
		{
			name: "base-middleware",
			routes: []*route.Route{
				{
					Path: "/",
					Middleware: []mux.MiddlewareFunc{
						func(next http.Handler) http.Handler {
							return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
								next.ServeHTTP(w, r)
							})
						},
					},
					Subroutes: []*route.Route{
						{
							Path: "/a",
							Handler: func(w http.ResponseWriter, r *http.Request) {
								w.Write([]byte("OK"))
							},
						},
					},
				},
			},
			url: "/a",
		},
	}

	zapLog, _ := zap.NewDevelopment()
	logger := zapr.NewLogger(zapLog)

	defaultRoute := func(_ *Router) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			logger.Info("Unhandled route", "URL", r.URL)
			w.WriteHeader(404)
		}
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			rtr := New(logger)
			mux, err := rtr.Route(defaultRoute, tc.routes)
			if err != nil {
				t.Fatal("failed to configure routes")
			}

			ts := httptest.NewServer(mux)
			defer ts.Close()

			url := ts.URL + tc.url
			t.Logf("Have URL '%s'", url)

			res, status := get(t, url)
			if res != "OK" {
				t.Errorf("failed to get OK: got '%s'", res)
			}
			if status != http.StatusOK {
				t.Errorf("failed to get OK status: got '%d'", status)
			}
		})
	}
}

func get(t *testing.T, url string) (string, int) {
	res, err := http.Get(url)
	if err != nil {
		t.Errorf("failed to perform get")
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("failed to read body")
	}
	return string(body), res.StatusCode
}
