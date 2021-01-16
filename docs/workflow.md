# Workflow

## Observing a helm deployment

When a user triggers a helm deployment, helm first creates a secret in the deployment's namespace.

The `tugboat controller` watches a subset of namespaces for the creation of secrets via a validating webhook.  Note that the `tugboat` validating webbook will _never_ send a responce that a secret is not to be admitted.  If a secret is created with the type `TODO` and whose name matches the pattern `TODO`, then `tugboat` will create a `releasehistory` custom resource.

The `tugboat` controller keeps a cache of `releasehistory`.

After `helm` has created its release (as represented by the secret), it will start to create the k8s resources.  `tugboat` uses a mutating webhook to observe creation of the resources.  When a resource is created, `tugboat` applies a label, `TODO`, with the value `TODO`.

When the `helm install` and `helm upgrade` are run, it applies certain annotations and labels to the kubernetes resources:
```
metadata:
  annotations:
    meta.helm.sh/release-name: RELEASE-NAME
    meta.helm.sh/release-namespace: RELEASE-NAMESPACE
  labels:
    app.kubernetes.io/managed-by: Helm
```

These are not applied to pods that are created _by_ k8s resources that are controlled by `helm`.  For example, a `helm` chart that specifes a `deployment` will have the annotations and label, but the pods which are created by the deployment will not.

However, Kubernetes provides an `ownerReference` metadata field, which references one or more other resources.  If this object map traces back to a `helm`-owned resource, then `tugboat` should track it.




The `tugboat watcher`, meanwhile, keeps a cache of `releasehistory` custom resources.  The `releasehistory` tracks the state of the helm deployment.  

It uses the kubernetes watcher pattern to observe all resources within a specified namespace or namespaces.  It will store events in the `releasehistory` status, such as "pod created", "deployment modified", etc.  Additionally, it will store

When the `tugboat watcher` observes that something interesting has happened, it sends a message to all its registered notifiers.  The notifiers then inform the user of these events, as approprite.  The `notifier-slack` may instantly send a message to some specified Slack channel when a deployment starts, while a `notifier-email` may wait until the deployment is _complete_ a send a single summary of events.