# Tugboat

_Tugboat is a work in progress, and documentation here does not reflect the current state of the project._

Tugboat improves the [Helm](helm.sh) chart deployment process for developers by exposing pod state and state transitions during a deployment.

In some kubernetes environments, a software developer may not have complete access all Kubernetes resources. For example, a production environment may be walled off. Tools such as Splunk or Datadog give an excellent _overview_ of a complete Kubernetes environment, but may not provide an easily accessible, _acute_ view of changes relevant to a deployment. This is not a criticism of these tools, just an observation of the space that Tugboat intends to fill.

## What questions does Tugboat answer?

During deployment...
* Have any of my pods updated?
* Have _all_ of my pods updated?
* How are the new pods configured (docker image, etc.)?
* Are my pods in a Crashloop Backoff?

## Do I need Tugboat?

In short, no. Tugboat does not provide any mission critical tooling not already made available through other means. Tugboat exists as a _facilitator_ for SREs and developers, providing proactive insight into existent data.

You may want Tugboat if an environment is unavailable via normal `kubectl` or `helm` commands, log collection is far from pod state, or state is not represented sufficiently quickly.

## Why the name "Tugboat"?

The Kubernetes ecosystem follows a nautical theme (note the Kubernetes logo, applications such as `helm`, `spinnaker`, etc.). A [tugboat](https://en.wikipedia.org/wiki/Tugboat) (the nautical vessel) assists other, large boats as they launch from a port, ensuring that they are able to manuever into open waters.

While the process is not completely analogous, Tugboat (the software) is designed to help deploy software by hooking into a Kubernetes environment and providing feedback to a developer about what is happening during the deployment itself, in relatively real-time.

## Limitations

Tugboat is not intended to observe the state of Kubernetes resources such as nodes.

Documentation:
*  for Tugboat developers
  * [Building](docs/development.md)
  * [Running locally](docs/running-locally.md)