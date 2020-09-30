// +build !test

package tools

//go:generate go build -o ../bin/protoc-gen-go ../vendor/google.golang.org/protobuf/cmd/protoc-gen-go
//go:generate go build -o ../bin/protoc-gen-go-grpc ../vendor/google.golang.org/grpc/cmd/protoc-gen-go-grpc
//go:generate mkdir -p ../internal/generated/notifier
//go:generate protoc --proto_path=../internal/proto/notifier --go_opt=paths=source_relative --go_out=../internal/generated/notifier --go-grpc_opt=paths=source_relative --go-grpc_out=../internal/generated/notifier ../internal/proto/notifier/notify.proto
