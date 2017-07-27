// Package otbuiltin contains all of the basic commands for creating and
// interacting with an ostree repository
package otbuiltin

import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"

	glib "github.com/ostreedev/ostree-go/pkg/glibobject"
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

func cCancellable(c *glib.GCancellable) *C.GCancellable {
	return (*C.GCancellable)(unsafe.Pointer(c.Native()))
}

// Converts an ostree repo struct to its C equivalent
func (r *Repo) native() *C.OstreeRepo {
	//return (*C.OstreeRepo)(r.Ptr())
	return (*C.OstreeRepo)(r.ptr)
}

// Takes a C ostree repo and converts it to a Go struct
func repoFromNative(p *C.OstreeRepo) *Repo {
	if p == nil {
		return nil
	}
	//o := (*glib.GObject)(unsafe.Pointer(p))
	//r := &Repo{o}
	r := &Repo{unsafe.Pointer(p)}
	return r
}

// Checks if the repo has been initialized
func (r *Repo) isInitialized() bool {
	if r.ptr != nil {
		return true
	}
	return false
}

func (repo *Repo) ResolveRev(refspec string, allowNoent bool) (string, error) {
	var cerr *C.GError = nil
	var coutrev *C.char = nil

	crefspec := C.CString(refspec)
	defer C.free(unsafe.Pointer(crefspec))

	r := C.ostree_repo_resolve_rev(repo.native(), crefspec, (C.gboolean)(glib.GBool(allowNoent)), &coutrev, &cerr)
	if !gobool(r) {
		return "", generateError(cerr)
	}

	outrev := C.GoString(coutrev)
	C.free(unsafe.Pointer(coutrev))

	return outrev, nil
}

// Attempts to open the repo at the given path
func OpenRepo(path string) (*Repo, error) {
	var cerr *C.GError = nil
	cpath := C.CString(path)
	pathc := C.g_file_new_for_path(cpath)
	defer C.g_object_unref(C.gpointer(pathc))
	crepo := C.ostree_repo_new(pathc)
	repo := repoFromNative(crepo)
	r := glib.GoBool(glib.GBoolean(C.ostree_repo_open(crepo, nil, &cerr)))
	if !r {
		return nil, generateError(cerr)
	}
	return repo, nil
}

type PullOptions struct {
	OverrideRemoteName string
	Refs               []string
}

func (repo *Repo) PullWithOptions(remoteName string, options PullOptions, progress *AsyncProgress, cancellable *glib.GCancellable) error {
	var cerr *C.GError = nil

	cremoteName := C.CString(remoteName)
	defer C.free(unsafe.Pointer(cremoteName))

	builder := C.g_variant_builder_new(C._g_variant_type(C.CString("a{sv}")))
	if options.OverrideRemoteName != "" {
		cstr := C.CString(options.OverrideRemoteName)
		v := C.g_variant_new_take_string((*C.gchar)(cstr))
		k := C.CString("override-remote-name")
		defer C.free(unsafe.Pointer(k))
		C._g_variant_builder_add_twoargs(builder, C.CString("{sv}"), k, v)
	}

	if len(options.Refs) != 0 {
		crefs := make([]*C.gchar, len(options.Refs))
		for i, s := range options.Refs {
			crefs[i] = (*C.gchar)(C.CString(s))
		}

		v := C.g_variant_new_strv((**C.gchar)(&crefs[0]), (C.gssize)(len(crefs)))

		for i, s := range crefs {
			crefs[i] = nil
			C.free(unsafe.Pointer(s))
		}

		k := C.CString("refs")
		defer C.free(unsafe.Pointer(k))

		C._g_variant_builder_add_twoargs(builder, C.CString("{sv}"), k, v)
	}

	coptions := C.g_variant_builder_end(builder)

	r := C.ostree_repo_pull_with_options(repo.native(), cremoteName, coptions, progress.native(), cCancellable(cancellable), &cerr)

	if !gobool(r) {
		return generateError(cerr)
	}

	return nil
}

// Enable support for tombstone commits, which allow the repo to distinguish between
// commits that were intentionally deleted and commits that were removed accidentally
func enableTombstoneCommits(repo *Repo) error {
	var tombstoneCommits bool
	var config *C.GKeyFile = C.ostree_repo_get_config(repo.native())
	var cerr *C.GError

	tombstoneCommits = glib.GoBool(glib.GBoolean(C.g_key_file_get_boolean(config, (*C.gchar)(C.CString("core")), (*C.gchar)(C.CString("tombstone-commits")), nil)))

	//tombstoneCommits is false only if it really is false or if it is set to FALSE in the config file
	if !tombstoneCommits {
		C.g_key_file_set_boolean(config, (*C.gchar)(C.CString("core")), (*C.gchar)(C.CString("tombstone-commits")), C.TRUE)
		if !glib.GoBool(glib.GBoolean(C.ostree_repo_write_config(repo.native(), config, &cerr))) {
			return generateError(cerr)
		}
	}
	return nil
}

func generateError(err *C.GError) error {
	goErr := glib.ConvertGError(glib.ToGError(unsafe.Pointer(err)))
	_, file, line, ok := runtime.Caller(1)
	if ok {
		return errors.New(fmt.Sprintf("%s:%d - %s", file, line, goErr))
	} else {
		return goErr
	}
}

func gobool(b C.gboolean) bool {
	return b != C.FALSE
}

type AsyncProgress struct {
	*glib.Object
}

func (a *AsyncProgress) native() *C.OstreeAsyncProgress {
	if a == nil {
		return nil
	}
	return (*C.OstreeAsyncProgress)(unsafe.Pointer(a.Native()))
}
