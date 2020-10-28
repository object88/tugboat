// +build test

package tools

//go:generate go build -o ../bin/mockgen ../vendor/github.com/golang/mock/mockgen
//go:generate mkdir -p ../mocks
//go:generate ../bin/mockgen -destination=../mocks/mock_grpc.go -package=mocks github.com/object88/tugboat/pkg/grpc/server Handler
//go:generate ../bin/mockgen -destination=../mocks/mock_httproundtripper.go -package=mocks net/http RoundTripper
//go:generate ../bin/mockgen -destination=../mocks/mock_validator_webhookprocessor.go -package=mocks github.com/object88/tugboat/apps/tugboat-controller/pkg/validator WebhookProcessor
//go:generate mkdir -p ../mocks/grpc/sample
//go:generate protoc --proto_path=../testdata/proto/sample --go_opt=paths=source_relative --go_out=../mocks/grpc/sample --go-grpc_opt=paths=source_relative --go-grpc_out=../mocks/grpc/sample ../testdata/proto/sample/sample.proto
