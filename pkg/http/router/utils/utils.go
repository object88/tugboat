package utils

import (
	"fmt"
	"net/http"

	"github.com/Masterminds/semver/v3"
	"github.com/gorilla/mux"
)

func ReadQueryParam(w http.ResponseWriter, r *http.Request, name string) (string, bool) {
	value, ok := mux.Vars(r)[name]
	if !ok {
		s := fmt.Sprintf("Missing query param '%s'", name)
		w.Write([]byte(s))
		w.WriteHeader(http.StatusBadRequest)
		return "", false
	}
	return value, true
}

func ReadVersionQueryParam(w http.ResponseWriter, r *http.Request, name string) (*semver.Version, bool) {
	raw, ok := ReadQueryParam(w, r, name)
	if !ok {
		return nil, false
	}
	ver, err := semver.StrictNewVersion(raw)
	if err != nil {
		// TODO: Replace this logger -- maybe use a context to store the provider
		// rtr.prov.Logger().Infof("Failed to parse version string '%s'", raw)
		return nil, false
	}
	return ver, true
}
