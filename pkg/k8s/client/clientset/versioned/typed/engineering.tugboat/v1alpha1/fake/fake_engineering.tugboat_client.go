/*
LICENSE
*/
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/object88/tugboat/pkg/k8s/client/clientset/versioned/typed/engineering.tugboat/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeTugboatV1alpha1 struct {
	*testing.Fake
}

func (c *FakeTugboatV1alpha1) ReleaseHistories(namespace string) v1alpha1.ReleaseHistoryInterface {
	return &FakeReleaseHistories{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeTugboatV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}