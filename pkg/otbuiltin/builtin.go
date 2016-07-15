package otbuiltin

import (
	"unsafe"

	glib "github.com/14rcole/ostree-go/pkg/glibobject"
)

// #cgo pkg-config: ostree-1
// #include <stdlib.h>
// #include <glib.h>
// #include <ostree.h>
// #include "builtin.go.h"
import "C"

type Repo struct {
	//*glib.GObject
	ptr unsafe.Pointer
}

func (r *Repo) native() *C.OstreeRepo {
	//return (*C.OstreeRepo)(r.Ptr())
	return (*C.OstreeRepo)(r.ptr)
}

func repoFromNative(p *C.OstreeRepo) *Repo {
	if p == nil {
		return nil
	}
	//o := (*glib.GObject)(unsafe.Pointer(p))
	//r := &Repo{o}
	r := &Repo{unsafe.Pointer(p)}
	return r
}

func (r *Repo) isInitialized() bool {
	if r.ptr != nil {
		return true
	}
	return false
}

func openRepo(path string) (*Repo, error) {
	var cerr *C.GError = nil
	cpath := C.CString(path)
	pathc := C.g_file_new_for_path(cpath)
	defer C.g_object_unref(C.gpointer(pathc))
	crepo := C.ostree_repo_new(pathc)
	repo := repoFromNative(crepo)
	r := glib.GoBool(glib.GBoolean(C.ostree_repo_open(crepo, nil, &cerr)))
	if !r {
		return nil, glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
	}
	return repo, nil
}

func enableTombstoneCommits(repo *Repo) error {
	var tombstoneCommits bool
	var config *C.GKeyFile = C.ostree_repo_get_config(repo.native())
	var cerr *C.GError

	tombstoneCommits = glib.GoBool(glib.GBoolean(C.g_key_file_get_boolean(config, (*C.gchar)(C.CString("core")), (*C.gchar)(C.CString("tombstone-commits")), nil)))

	//tombstoneCommits is false only if it really is false or if it is set to FALSE in the config file
	if !tombstoneCommits {
		C.g_key_file_set_boolean(config, (*C.gchar)(C.CString("core")), (*C.gchar)(C.CString("tombstone-commits")), C.TRUE)
		if !glib.GoBool(glib.GBoolean(C.ostree_repo_write_config(repo.native(), config, &cerr))) {
			return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
		}
	}
	return nil
}
