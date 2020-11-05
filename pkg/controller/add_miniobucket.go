package controller

import (
	"github.com/Walkbase/minio-resources-operator/pkg/controller/miniobucket"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, miniobucket.Add)
}
