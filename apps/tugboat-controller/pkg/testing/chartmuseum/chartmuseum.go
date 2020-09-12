package chartmuseum

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/chartmuseum/storage"
	"github.com/gin-gonic/gin"
	"helm.sh/chartmuseum/pkg/chartmuseum/logger"
	"helm.sh/chartmuseum/pkg/chartmuseum/router"
	"helm.sh/chartmuseum/pkg/chartmuseum/server/multitenant"
)

type TestChartMuseum struct {
	Server *multitenant.MultiTenantServer

	HTTPServer *httptest.Server

	Requests map[string]int
}

func NewTestChartMuseum(storagepath string) (*TestChartMuseum, error) {
	tcm := &TestChartMuseum{
		Requests: map[string]int{},
	}

	lggr, err := logger.NewLogger(logger.LoggerOptions{
		Debug:   false,
		LogJSON: true,
	})
	if err != nil {
		return nil, fmt.Errorf("Internal error: failed to create chart museum logger: %w", err)
	}

	rtr := router.NewRouter(router.RouterOptions{
		Logger:        lggr,
		ContextPath:   "",
		AnonymousGet:  true,
		Depth:         0,
		MaxUploadSize: 1024 * 512,
	})

	rtr.Use(tcm.Track)

	storageBackend := storage.NewLocalFilesystemBackend(storagepath)

	options := multitenant.MultiTenantServerOptions{
		Logger:         lggr,
		Router:         rtr,
		StorageBackend: storageBackend,
		IndexLimit:     1,
		EnableAPI:      true,
		DisableDelete:  true,
	}

	tcm.Server, err = multitenant.NewMultiTenantServer(options)
	if err != nil {
		return nil, fmt.Errorf("Internal error: failed to create multi-tenant server: %w", err)
	}

	return tcm, nil
}

func (tcm *TestChartMuseum) Run() {
	tcm.HTTPServer = httptest.NewUnstartedServer(tcm.Server.Router)
	tcm.HTTPServer.Start()
}

func (tcm *TestChartMuseum) Close() error {
	tcm.HTTPServer.Close()
	return nil
}

func (tcm *TestChartMuseum) RefreshIndex() error {
	client := tcm.HTTPServer.Client()
	resp, err := client.Get(tcm.HTTPServer.URL + "/index.yaml")
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()
	if err != nil {
		return fmt.Errorf("internal error: failed to get index.yaml: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Expected OK, got %d/%s: %s", resp.StatusCode, resp.Status, string(body))
	}

	return nil
}

func (tcm *TestChartMuseum) UploadArchive(packagePath string) error {
	f, err := os.Open(packagePath)
	if err != nil {
		return fmt.Errorf("internal error: failed to upload chart: %w", err)
	}
	client := tcm.HTTPServer.Client()
	resp, err := client.Post(tcm.HTTPServer.URL+"/api/charts", "application/tar+gzip", f)
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()
	if err != nil {
		return fmt.Errorf("internal error: failed to post tarball to museum: %w", err)
	}
	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("expected Created, got %d/%s: %s", resp.StatusCode, resp.Status, string(body))
	}

	return nil
}

func (tcm *TestChartMuseum) Track(c *gin.Context) {
	if i, ok := tcm.Requests[c.Request.URL.Path]; !ok {
		tcm.Requests[c.Request.URL.Path] = 1
	} else {
		tcm.Requests[c.Request.URL.Path] = i + 1
	}
	c.Next()
}

func (tcm *TestChartMuseum) DumpRequests() {
	for k, v := range tcm.Requests {
		fmt.Printf("%s: %d\n", k, v)
	}
}
