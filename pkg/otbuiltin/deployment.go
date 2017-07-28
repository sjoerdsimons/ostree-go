package otbuiltin

import (
	"runtime"
	"unsafe"

	glib "github.com/sjoerdsimons/ostree-go/pkg/glibobject"
)

// #cgo pkg-config: ostree-1
// #include <stdlib.h>
// #include <glib.h>
// #include <ostree.h>
// #include "builtin.go.h"
import "C"

type Deployment struct {
	*glib.Object
}

func wrapDeployment(d *C.OstreeDeployment) *Deployment {
	g := glib.ToGObject(unsafe.Pointer(d))
	obj := &glib.Object{g}
	deployment := &Deployment{obj}

	runtime.SetFinalizer(deployment, (*Deployment).Unref)

	return deployment
}

func (d *Deployment) native() *C.OstreeDeployment {
	if d == nil || d.GObject == nil {
		return nil
	}
	return (*C.OstreeDeployment)(unsafe.Pointer(d.GObject))
}

func (d *Deployment) Osname() string {
	return C.GoString(C.ostree_deployment_get_osname(d.native()))
}

func (d *Deployment) Csum() string {
	return C.GoString(C.ostree_deployment_get_csum(d.native()))
}

func (d *Deployment) Deployserial() int {
	return int(C.ostree_deployment_get_deployserial(d.native()))
}

func (d *Deployment) OriginRelpath() string {
	cstr := C.ostree_deployment_get_origin_relpath(d.native())

	str := C.GoString(cstr)
	C.free(unsafe.Pointer(cstr))

	return str
}
