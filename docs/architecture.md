# Architecture

Tugboat consists of two primary components: `tugboat-controller` and `tugboat-watcher`.

## ReleaseHistory

The ReleaseHistory custom resource (`releasehistories.tugboat.engineering`) encapsulates the historic record of the kubernetes resources described by a Helm deployment.  It lives in the same namespace as the helm release itself.

Spec:
| Property | Type | Required | Description |
| --- | --- | --- | --- |
| `releasename` | string | Y | The name of the Helm release |

Status
| Property | Type | Description |
| --- | --- | --- |
| `active` | bool | Indicates whether the chart is still deployed
| `events` | []Event | 

Event
| Property | Type | Description |
| --- | --- | --- |


## Tugboat Controller

The tugboat controller manages `releasehistories.tugboat.engineering` custom resources. The tugboat controller runs within the cluster that it observes.

### Validating input

`tugboat-controller` uses a [Validating Admission Webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/) to ensure the correctness of incoming `launches`. Once a `launch` has been created, some fields cannot be changed, such as the chart, while others can, such as the chart _version_. But it is also important that the chart version is published and accessible. A validating admission webhook can [address these concerns](https://www.openshift.com/blog/kubernetes-operators-best-practices); once past the webhook, the resource is written into `etcd` (or other storage), and the controller itself will have to deal with any illegal state.

Reference material for further understanding of Validating Admission Webbooks:
* [In-depth introduction to Kubernetes admission webhooks](https://banzaicloud.com/blog/k8s-admission-webhooks/)
* [Diving into Kubernetes Mutating Webhooks](https://medium.com/ibm-cloud/diving-into-kubernetes-mutatingadmissionwebhook-6ef3c5695f74)
* [Writing a very basic kubernetes mutating admission webhook](https://medium.com/ovni/writing-a-very-basic-kubernetes-mutating-admission-webhook-398dbbcb63ec)

Note that the current implementation deployed with a _self-signed certificate_, and should not be put into production.

## Tugboat Watcher


# Notes

How to track objects types as they are in scope and out of scope.

* Mutator tracks resource declarations as they appear
* 