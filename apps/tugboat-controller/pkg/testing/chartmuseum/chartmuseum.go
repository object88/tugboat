package chartmuseum

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/object88/tugboat/pkg/http/router"
	"github.com/object88/tugboat/pkg/http/router/route"
	"github.com/object88/tugboat/pkg/http/router/utils"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"
)

// TestChartMuseum is a in-memory chart museum with extremely limited
// capabilities: it can return a tarball and the index.yaml; no other API is
// provided
type TestChartMuseum struct {
	Logger logr.Logger

	HTTPServer *httptest.Server
	Router     *mux.Router
	Contents   map[string][]byte
	Entries    map[string]repo.ChartVersions

	Requests map[string]int
}

// NewTestChartMuseum returns a new instance of the TestChartMuseum struct
func NewTestChartMuseum(logger logr.Logger) (*TestChartMuseum, error) {
	tcm := &TestChartMuseum{
		Logger:   logger,
		Entries:  map[string]repo.ChartVersions{},
		Requests: map[string]int{},
		Contents: map[string][]byte{},
	}

	defaultRoute := func(rtr *router.Router) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			rtr.Logger.Info("Unhandled URL", "URL", r.URL)
			w.WriteHeader(404)
		}
	}

	routes := []*route.Route{
		{
			Path:       "/",
			Middleware: []mux.MiddlewareFunc{tcm.Track()},
			Subroutes: []*route.Route{
				{
					Path: "/",
					Handler: func(w http.ResponseWriter, r *http.Request) {
						io.WriteString(w, "test chart museum")
					},
					Methods: []string{http.MethodGet},
				},
				{
					Path:    "/index.yaml",
					Handler: tcm.HandleIndex,
					Methods: []string{http.MethodGet},
				},
				{
					Path:    "/charts/{tarball}",
					Handler: tcm.HandleTarball,
					Methods: []string{http.MethodGet},
				},
			},
		},
	}

	var err error
	tcm.Router, err = router.New(tcm.Logger).Route(defaultRoute, routes)
	if err != nil {
		return nil, err
	}

	return tcm, nil
}

func (tcm *TestChartMuseum) HandleIndex(w http.ResponseWriter, r *http.Request) {
	// Sample:
	// apiVersion: v1
	// entries:
	// 	superfoo:
	// 	- apiVersion: v2
	// 		appVersion: 0.1.0
	// 		created: "2020-10-03T20:29:35.28684638-07:00"
	// 		description: test chart superfoo
	// 		digest: ed7b69a5420e10558e94696c61cf3802946ba490fd7a508b54c27bbbdfc48256
	// 		name: superfoo
	// 		type: application
	// 		urls:
	// 		- charts/superfoo-1.2.3.tgz
	// 		version: 1.2.3
	// generated: "2020-10-03T20:29:35-07:00"
	// serverInfo: {}

	indexFile := repo.NewIndexFile()
	indexFile.Entries = tcm.Entries

	if buf, err := yaml.Marshal(indexFile); err != nil {
		w.WriteHeader(400)
	} else {
		w.Write(buf)
	}
}

func (tcm *TestChartMuseum) HandleTarball(w http.ResponseWriter, r *http.Request) {
	if raw, ok := utils.ReadQueryParam(w, r, "tarball"); !ok {
		w.WriteHeader(500)
	} else if contents, ok := tcm.Contents[raw]; !ok {
		tcm.Logger.Info("did not find tarball", "tarball", raw)
		w.WriteHeader(500)
	} else {
		tcm.Logger.Info("going to write", "tarball", raw, "bytecount", len(contents))
		w.Write(contents)
	}
}

// Run starts the test HTTP server
func (tcm *TestChartMuseum) Run() {
	tcm.HTTPServer = httptest.NewUnstartedServer(tcm.Router)
	tcm.HTTPServer.Start()
}

// Close implements io.Closer and will stop the HTTP server
func (tcm *TestChartMuseum) Close() error {
	tcm.HTTPServer.Close()
	return nil
}

// UploadArchive injects a chart into the chart museum
func (tcm *TestChartMuseum) UploadArchive(chartname string, chartversion string, packagePath string) error {
	buf, err := ioutil.ReadFile(packagePath)
	if err != nil {
		return err
	}

	tarballname := fmt.Sprintf("%s-%s.tgz", chartname, chartversion)

	tcm.Logger.Info("Have buffer", "bytecount", len(buf), "tarball", tarballname)

	newversion := &repo.ChartVersion{
		Metadata: &chart.Metadata{
			APIVersion: chart.APIVersionV2,
			Name:       chartname,
			Version:    chartversion,
		},
		URLs: []string{fmt.Sprintf("/charts/%s", tarballname)},
	}

	tcm.Contents[filepath.Base(packagePath)] = buf
	versions, ok := tcm.Entries[chartname]
	if !ok {
		tcm.Entries[chartname] = []*repo.ChartVersion{newversion}
	} else {
		tcm.Entries[chartname] = append(versions, newversion)
	}

	return nil
}

func (tcm *TestChartMuseum) Track() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tcm.Logger.Info("Req", "URL", r.URL)
			p := r.URL.Path
			if i, ok := tcm.Requests[p]; !ok {
				tcm.Requests[p] = 1
			} else {
				tcm.Requests[p] = i + 1
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (tcm *TestChartMuseum) DumpRequests() {
	for k, v := range tcm.Requests {
		tcm.Logger.Info("", "req", k, "count", v)
	}
}
