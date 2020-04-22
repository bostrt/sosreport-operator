package controller

import (
	"github.com/bostrt/sosreport-operator/pkg/controller/sosreport"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, sosreport.Add)
}
