package controller

import (
	"github.com/robotinfra/minio-resources-operator/pkg/controller/minioserver"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, minioserver.Add)
}
