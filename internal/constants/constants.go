package constants

import v1 "k8s.io/api/core/v1"

const (
	HelmSecretFinalizer string = "engineering.tugboat/helm-secret-finalizer"
)

const (
	HelmSecretType v1.SecretType = "helm.sh/release.v1"
)
