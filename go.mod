module github.com/object88/tugboat

go 1.15

require (
	github.com/Masterminds/semver/v3 v3.1.0
	github.com/emicklei/go-restful v2.11.1+incompatible // indirect
	github.com/go-logr/logr v0.2.1
	github.com/go-logr/zapr v0.2.0
	github.com/go-openapi/spec v0.19.6 // indirect
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.1.2
	github.com/googleapis/gnostic v0.5.1 // indirect
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/gregjones/httpcache v0.0.0-20181110185634-c63ab54fda8f // indirect
	github.com/hashicorp/go-multierror v1.1.0
	github.com/slack-go/slack v0.7.2
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1 // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/net v0.0.0-20200813134508-3edf25e44fcc // indirect
	google.golang.org/grpc v1.33.1
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.0.0
	google.golang.org/protobuf v1.25.0
	k8s.io/api v0.19.3
	k8s.io/apiextensions-apiserver v0.19.3
	k8s.io/apimachinery v0.19.3
	k8s.io/cli-runtime v0.19.3
	k8s.io/client-go v0.19.3
	k8s.io/code-generator v0.19.3
	sigs.k8s.io/controller-runtime v0.6.3
)

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8
