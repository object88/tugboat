package cache

import (
	"context"
	"testing"

	"github.com/go-logr/zapr"
	"github.com/golang/mock/gomock"
	"github.com/object88/tugboat/mocks"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
)

func Test_Cache_Hit(t *testing.T) {
	chartName := "foo"
	chartVersion := "1.2.3"
	repoName := "repo"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mrl := mocks.NewMockRepoLoader(ctrl)

	// Do not establish any mocked calls; the cache should not need to make any
	// such request.

	repository := repo.Entry{
		Name: repoName,
	}

	c := Cache{}
	c.Initialize(mrl)
	c.AddRepository(&repository)
	c.contents[repoName].contents[chartName] = &countedrepo{
		contents: map[string]*countedversion{
			chartVersion: {
				cv: &repo.ChartVersion{
					Metadata: &chart.Metadata{},
				},
			},
		},
	}

	ctx := context.Background()
	m, ok, err := c.Get(ctx, repoName, chartName, chartVersion)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if !ok {
		t.Errorf("Did not get a metadata")
	}
	if m == nil {
		t.Errorf("Did not get *chart.Metadata")
	}
}

func Test_Cache_Get_Miss(t *testing.T) {
	chartName := "foo"
	chartVersion := "1.2.3"
	repoName := "repo"

	repository := repo.Entry{
		Name: repoName,
	}

	tcs := []struct {
		name     string
		preloads []preload
	}{
		{
			name:     "empty cache",
			preloads: []preload{},
		},
		{
			name:     "cache with different chart name",
			preloads: []preload{{name: "quux", version: chartVersion}},
		},
		{
			name:     "cache with different chart version",
			preloads: []preload{{name: chartName, version: "2.3.4"}},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mrl := mocks.NewMockRepoLoader(ctrl)

			zapLog, _ := zap.NewDevelopment()
			logger := zapr.NewLogger(zapLog)
			c := Cache{
				Config: Config{
					Logger: logger,
				},
			}
			c.Initialize(mrl)
			c.AddRepository(&repository)
			c.Load(repoName, ToIndexFile(tc.preloads...))

			// Prepare for the miss fetch
			idx := repo.IndexFile{
				Entries: map[string]repo.ChartVersions{
					chartName: {
						&repo.ChartVersion{
							Metadata: &chart.Metadata{
								Name:    chartName,
								Version: chartVersion,
							},
						},
					},
				},
			}
			mrl.EXPECT().GetRepoIndexFile(repoName).Return(&idx, nil).Times(1)

			// Run the test
			m, ok, err := c.Get(context.Background(), repoName, chartName, chartVersion)
			if err != nil {
				t.Errorf("Unexpected error: %s", err.Error())
			}
			if !ok {
				t.Errorf("NOTOK")
			}
			if m == nil {
				t.Errorf("Did not get *chart.Metadata")
			}

			// Check the internal store
			if _, ok := c.contents[repoName].contents[chartName].contents[chartVersion]; !ok {
				t.Errorf("Chart not stored")
			}
		})
	}
}

type preload struct {
	name    string
	version string
}

func ToIndexFile(preloads ...preload) *repo.IndexFile {
	entries := map[string]repo.ChartVersions{}

	for _, p := range preloads {
		versions, ok := entries[p.name]
		cv := &repo.ChartVersion{
			Metadata: &chart.Metadata{
				Name:    p.name,
				Version: p.version,
			},
		}
		if !ok {
			versions = repo.ChartVersions{cv}
		} else {
			versions = append(versions, cv)
		}
		entries[p.name] = versions
	}

	return &repo.IndexFile{Entries: entries}
}
