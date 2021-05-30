# New Architecture

## Controller

The Controller managed a k8s controller that watches for secrets, filtered to match the pattern that `helm` uses.

## Watcher

The Watcher observes k8s events to follow the lifecycle of all k8s objects.  The watcher may be filtered down to specific namespaces or namespace patterns.