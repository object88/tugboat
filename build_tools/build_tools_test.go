package tools

//go:generate go build -o ../bin/mockgen ../vendor/github.com/golang/mock/mockgen
//go:generate mkdir -p ../mocks

//go:generate ../bin/mockgen -destination=../mocks/mock_helm_loader.go -package=mocks github.com/object88/tugboat/apps/tugboat-controller/pkg/helm/cache/repos/loader RepoLoader
