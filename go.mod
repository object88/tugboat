module github.com/object88/tugboat

go 1.15

require (
	github.com/MakeNowJust/heredoc v0.0.0-20171113091838-e9091a26100e // indirect
	github.com/bugsnag/bugsnag-go v1.5.3 // indirect
	github.com/bugsnag/panicwrap v1.2.0 // indirect
	// github.com/coreos/etcd v3.3.15+incompatible // indirect
	github.com/docker/go-metrics v0.0.0-20181218153428-b84716841b82 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/emicklei/go-restful v2.11.1+incompatible // indirect
	github.com/garyburd/redigo v1.6.2 // indirect
	github.com/go-openapi/spec v0.19.6 // indirect
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	// github.com/golang/lint v0.0.0-20180702182130-06c8688daad7 // indirect
	github.com/golang/mock v1.4.4
	github.com/googleapis/gnostic v0.5.1 // indirect
	github.com/gorilla/handlers v1.4.2 // indirect
	github.com/gregjones/httpcache v0.0.0-20181110185634-c63ab54fda8f // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	// github.com/xenolf/lego v0.3.2-0.20160613233155-a9d8cec0e656 // indirect
	github.com/yvasiyarov/go-metrics v0.0.0-20150112132944-c25f46c4b940 // indirect
	github.com/yvasiyarov/gorelic v0.0.6 // indirect
	golang.org/x/net v0.0.0-20200813134508-3edf25e44fcc // indirect
	// gonum.org/v1/netlib v0.0.0-20190331212654-76723241ea4e // indirect
	// gopkg.in/square/go-jose.v1 v1.1.2 // indirect
	gopkg.in/yaml.v2 v2.3.0
	helm.sh/helm/v3 v3.3.0
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/cli-runtime v0.18.8
	k8s.io/client-go v0.18.8
	rsc.io/letsencrypt v0.0.3 // indirect
// sigs.k8s.io/structured-merge-diff v1.0.2 // indirect
)

// github.com/Azure/go-autorest/autorest has different versions for the Go
// modules than it does for releases on the repository. Note the correct
// version when updating.
// github.com/Azure/go-autorest/autorest => github.com/Azure/go-autorest/autorest v0.9.0
replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
