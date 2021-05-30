package constants

import v1 "k8s.io/api/core/v1"

const (
	HelmSecretLabelName     string = "name"
	HelmSecretLabelRevision        = "version"

	HelmSecretFinalizer       string = "engineering.tugboat/helm-secret-finalizer"
	HelmLabelReleaseName             = "meta.helm.sh/release-name"
	HelmLabelReleaseNamespace        = "meta.helm.sh/release-namespace"

	LabelReleaseName      = "tugboat.engineering/release-name"
	LabelReleaseNamespace = "tugboat.engineering/release-namespace"
	LabelRevision         = "tugboat.engineering/revision"
	LabelState            = "tugboat.engineering/state"
	LabelStateActive      = "active"
	LabelStateUninstalled = "uninstalled"
)

const (
	HelmSecretType v1.SecretType = "helm.sh/release.v1"
)
