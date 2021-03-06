/*
LICENSE
*/
// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/object88/tugboat/pkg/k8s/apis/engineering.tugboat/v1alpha1"
	scheme "github.com/object88/tugboat/pkg/k8s/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ReleaseHistoriesGetter has a method to return a ReleaseHistoryInterface.
// A group's client should implement this interface.
type ReleaseHistoriesGetter interface {
	ReleaseHistories(namespace string) ReleaseHistoryInterface
}

// ReleaseHistoryInterface has methods to work with ReleaseHistory resources.
type ReleaseHistoryInterface interface {
	Create(ctx context.Context, releaseHistory *v1alpha1.ReleaseHistory, opts v1.CreateOptions) (*v1alpha1.ReleaseHistory, error)
	Update(ctx context.Context, releaseHistory *v1alpha1.ReleaseHistory, opts v1.UpdateOptions) (*v1alpha1.ReleaseHistory, error)
	UpdateStatus(ctx context.Context, releaseHistory *v1alpha1.ReleaseHistory, opts v1.UpdateOptions) (*v1alpha1.ReleaseHistory, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.ReleaseHistory, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.ReleaseHistoryList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ReleaseHistory, err error)
	ReleaseHistoryExpansion
}

// releaseHistories implements ReleaseHistoryInterface
type releaseHistories struct {
	client rest.Interface
	ns     string
}

// newReleaseHistories returns a ReleaseHistories
func newReleaseHistories(c *TugboatV1alpha1Client, namespace string) *releaseHistories {
	return &releaseHistories{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the releaseHistory, and returns the corresponding releaseHistory object, and an error if there is any.
func (c *releaseHistories) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.ReleaseHistory, err error) {
	result = &v1alpha1.ReleaseHistory{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("releasehistories").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ReleaseHistories that match those selectors.
func (c *releaseHistories) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ReleaseHistoryList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.ReleaseHistoryList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("releasehistories").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested releaseHistories.
func (c *releaseHistories) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("releasehistories").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a releaseHistory and creates it.  Returns the server's representation of the releaseHistory, and an error, if there is any.
func (c *releaseHistories) Create(ctx context.Context, releaseHistory *v1alpha1.ReleaseHistory, opts v1.CreateOptions) (result *v1alpha1.ReleaseHistory, err error) {
	result = &v1alpha1.ReleaseHistory{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("releasehistories").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(releaseHistory).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a releaseHistory and updates it. Returns the server's representation of the releaseHistory, and an error, if there is any.
func (c *releaseHistories) Update(ctx context.Context, releaseHistory *v1alpha1.ReleaseHistory, opts v1.UpdateOptions) (result *v1alpha1.ReleaseHistory, err error) {
	result = &v1alpha1.ReleaseHistory{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("releasehistories").
		Name(releaseHistory.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(releaseHistory).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *releaseHistories) UpdateStatus(ctx context.Context, releaseHistory *v1alpha1.ReleaseHistory, opts v1.UpdateOptions) (result *v1alpha1.ReleaseHistory, err error) {
	result = &v1alpha1.ReleaseHistory{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("releasehistories").
		Name(releaseHistory.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(releaseHistory).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the releaseHistory and deletes it. Returns an error if one occurs.
func (c *releaseHistories) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("releasehistories").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *releaseHistories) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("releasehistories").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched releaseHistory.
func (c *releaseHistories) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ReleaseHistory, err error) {
	result = &v1alpha1.ReleaseHistory{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("releasehistories").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
