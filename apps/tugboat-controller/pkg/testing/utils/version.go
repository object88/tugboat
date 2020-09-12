package utils

import "github.com/Masterminds/semver/v3"

func WillMakeVersion(in string) *semver.Version {
	v, _ := semver.NewVersion(in)
	return v
}
