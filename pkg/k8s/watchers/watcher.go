package watchers

import (
	"k8s.io/client-go/tools/cache"
)

type Watcher interface {
	GetInformer() cache.SharedIndexInformer
}
