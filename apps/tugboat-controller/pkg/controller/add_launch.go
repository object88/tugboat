package controller

import (
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/controller/launch"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, launch.Add)
}
