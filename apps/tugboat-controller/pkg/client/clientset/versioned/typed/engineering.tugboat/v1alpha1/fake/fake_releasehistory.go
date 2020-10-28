/*
LICENSE
*/
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/object88/tugboat/apps/tugboat-controller/pkg/apis/engineering.tugboat/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeReleaseHistories implements ReleaseHistoryInterface
type FakeReleaseHistories struct {
	Fake *FakeTugboatV1alpha1
	ns   string
}

var releasehistoriesResource = schema.GroupVersionResource{Group: "tugboat.engineering", Version: "v1alpha1", Resource: "releasehistories"}

var releasehistoriesKind = schema.GroupVersionKind{Group: "tugboat.engineering", Version: "v1alpha1", Kind: "ReleaseHistory"}

// Get takes name of the releaseHistory, and returns the corresponding releaseHistory object, and an error if there is any.
func (c *FakeReleaseHistories) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.ReleaseHistory, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(releasehistoriesResource, c.ns, name), &v1alpha1.ReleaseHistory{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ReleaseHistory), err
}

// List takes label and field selectors, and returns the list of ReleaseHistories that match those selectors.
func (c *FakeReleaseHistories) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ReleaseHistoryList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(releasehistoriesResource, releasehistoriesKind, c.ns, opts), &v1alpha1.ReleaseHistoryList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ReleaseHistoryList{ListMeta: obj.(*v1alpha1.ReleaseHistoryList).ListMeta}
	for _, item := range obj.(*v1alpha1.ReleaseHistoryList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested releaseHistories.
func (c *FakeReleaseHistories) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(releasehistoriesResource, c.ns, opts))

}

// Create takes the representation of a releaseHistory and creates it.  Returns the server's representation of the releaseHistory, and an error, if there is any.
func (c *FakeReleaseHistories) Create(ctx context.Context, releaseHistory *v1alpha1.ReleaseHistory, opts v1.CreateOptions) (result *v1alpha1.ReleaseHistory, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(releasehistoriesResource, c.ns, releaseHistory), &v1alpha1.ReleaseHistory{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ReleaseHistory), err
}

// Update takes the representation of a releaseHistory and updates it. Returns the server's representation of the releaseHistory, and an error, if there is any.
func (c *FakeReleaseHistories) Update(ctx context.Context, releaseHistory *v1alpha1.ReleaseHistory, opts v1.UpdateOptions) (result *v1alpha1.ReleaseHistory, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(releasehistoriesResource, c.ns, releaseHistory), &v1alpha1.ReleaseHistory{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ReleaseHistory), err
}

// Delete takes name of the releaseHistory and deletes it. Returns an error if one occurs.
func (c *FakeReleaseHistories) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(releasehistoriesResource, c.ns, name), &v1alpha1.ReleaseHistory{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeReleaseHistories) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(releasehistoriesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.ReleaseHistoryList{})
	return err
}

// Patch applies the patch and returns the patched releaseHistory.
func (c *FakeReleaseHistories) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ReleaseHistory, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(releasehistoriesResource, c.ns, name, pt, data, subresources...), &v1alpha1.ReleaseHistory{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ReleaseHistory), err
}
