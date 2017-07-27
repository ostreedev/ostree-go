package otbuiltin

import (
	glib "github.com/ostreedev/ostree-go/pkg/glibobject"
	"unsafe"
)

// #cgo pkg-config: ostree-1
// #include <stdlib.h>
// #include <glib.h>
// #include <ostree.h>
// #include "builtin.go.h"
import "C"

func (repo *Repo) RemoteAdd(name, url string, options *glib.GVariant,
	cancellable *glib.GCancellable) error {

	var cerr *C.GError = nil

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))

	r := C.ostree_repo_remote_add(repo.native(), cname, curl, (*C.GVariant)(options.Ptr()), cCancellable(cancellable), &cerr)

	if !gobool(r) {
		return generateError(cerr)
	}

	return nil
}
