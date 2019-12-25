# Tugboat

Tugboat is a tool using to augment a [Helm](helm.sh) deployment by capturing the state of pods as their lifecycles end and start.

Requirements:
- pods must be labels with a release corresponding to the helm release name