package constants

import v1 "k8s.io/api/core/v1"

const (
	HelmSecretFinalizer       string = "engineering.tugboat/helm-secret-finalizer"
	HelmLabelReleaseName             = "meta.helm.sh/release-name"
	HelmLabelReleaseNamespace        = "meta.helm.sh/release-namespace"

	LabelReleaseName      = "tugboat.engineering/release-name"
	LabelReleaseNamespace = "tugboat.engineering/release-namespace"
	LabelState            = "tugboat.engineering/state"
	LabelStateActive      = "active"
	LabelStateUninstalled = "uninstalled"
)

const (
	HelmSecretType v1.SecretType = "helm.sh/release.v1"
)
