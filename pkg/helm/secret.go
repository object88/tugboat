package helm

import (
	"fmt"
	"regexp"
)

const (
	helmSecretNameRegex string = `^sh\.helm\.release\.v1\.(?P<name>.+)\.v(?P<revision>[1-9][0-9]*)$`
)

type SecretNameParser struct {
	reg                  *regexp.Regexp
	releaseNameIndex     int
	releaseRevisionIndex int
}

func New() (*SecretNameParser, error) {
	r, err := regexp.Compile(helmSecretNameRegex)
	if err != nil {
		return nil, fmt.Errorf("internal error; helm secret name regex failed to compile: %w", err)
	}
	nameindex := r.SubexpIndex("name")
	if nameindex == -1 {
		return nil, fmt.Errorf("internal error; failed to find 'releasename' subexp in helm secret name regex")
	}
	revisionindex := r.SubexpIndex("revision")

	w := &SecretNameParser{
		reg:                  r,
		releaseNameIndex:     nameindex,
		releaseRevisionIndex: revisionindex,
	}
	return w, nil
}

// func (snp SecretNameParser) Parse(in string) (string, int)
