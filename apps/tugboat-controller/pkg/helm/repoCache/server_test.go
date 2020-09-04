package repoCache

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/chartmuseum/storage"
	"helm.sh/chartmuseum/pkg/chartmuseum/logger"
	"helm.sh/chartmuseum/pkg/chartmuseum/router"
	"helm.sh/chartmuseum/pkg/chartmuseum/server/multitenant"
)

type testChartMuseum struct {
	server *multitenant.MultiTenantServer

	httpserver *httptest.Server
}

func newTestChartMuseum(storagepath string) (*testChartMuseum, error) {
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

	storageBackend := storage.NewLocalFilesystemBackend(storagepath)

	options := multitenant.MultiTenantServerOptions{
		Logger:         lggr,
		Router:         rtr,
		StorageBackend: storageBackend,
		IndexLimit:     1,
		EnableAPI:      true,
		DisableDelete:  true,
	}

	server, err := multitenant.NewMultiTenantServer(options)
	if err != nil {
		return nil, fmt.Errorf("Internal error: failed to create multi-tenant server: %w", err)
	}

	return &testChartMuseum{
		server: server,
	}, nil
}

func (tcm *testChartMuseum) Run() {
	tcm.httpserver = httptest.NewUnstartedServer(tcm.server.Router)
	tcm.httpserver.Start()
}

func (tcm *testChartMuseum) Close() error {
	tcm.httpserver.Close()
	return nil
}

func (tcm *testChartMuseum) refreshIndex() error {
	client := tcm.httpserver.Client()
	resp, err := client.Get(tcm.httpserver.URL + "/index.yaml")
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

func (tcm *testChartMuseum) uploadArchive(packagePath string) error {
	f, err := os.Open(packagePath)
	if err != nil {
		return fmt.Errorf("internal error: failed to upload chart: %w", err)
	}
	client := tcm.httpserver.Client()
	resp, err := client.Post(tcm.httpserver.URL+"/api/charts", "application/tar+gzip", f)
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
