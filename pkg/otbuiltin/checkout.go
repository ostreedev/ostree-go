package otbuiltin

import (
	"strings"
	"unsafe"

	glib "github.com/14rcole/ostree-go/pkg/glibobject"
)

// #cgo pkg-config: ostree-1
// #include <stdlib.h>
// #include <glib.h>
// #include <ostree.h>
// #include "builtin.go.h"
import "C"

var checkoutOpts checkoutOptions

type checkoutOptions struct {
	UserMode         bool   // Do not change file ownership or initialize extended attributes
	Union            bool   // Keep existing directories and unchanged files, overwriting existing filesystem
	AllowNoent       bool   // Do nothing if the specified filepath does not exist
	DisableCache     bool   // Do not update or use the internal repository uncompressed object caceh
	Whiteouts        bool   // Process 'whiteout' (docker style) entries
	RequireHardlinks bool   // Do not fall back to full copies if hard linking fails
	Subpath          string // Checkout sub-directory path
	FromFile         string // Process many checkouts from the given file
}

func NewCheckoutOptions() checkoutOptions {
	return checkoutOptions{}
}

func Checkout(repoPath, destination, commit string, opts checkoutOptions) error {
	checkoutOpts = opts

	var cancellable *glib.GCancellable
	ccommit := C.CString(commit)
	defer C.free(unsafe.Pointer(ccommit))
	var gerr = glib.NewGError()
	cerr := (*C.GError)(gerr.Ptr())
	defer C.free(unsafe.Pointer(cerr))

	repoPathc := C.g_file_new_for_path(C.CString(repoPath))
	defer C.g_object_unref(repoPathc)
	crepo := C.ostree_repo_new(repoPathc)
	if !glib.GoBool(glib.GBoolean(C.ostree_repo_open(crepo, (*C.GCancellable)(cancellable.Ptr()), &cerr))) {
		return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
	}

	if strings.Compare(checkoutOpts.FromFile, "") != 0 {
		err := processManyCheckouts(crepo, destination, cancellable)
		if err != nil {
			return err
		}
	} else {
		var resolvedCommit *C.char
		defer C.free(unsafe.Pointer(resolvedCommit))
		if !glib.GoBool(glib.GBoolean(C.ostree_repo_resolve_rev(crepo, ccommit, C.FALSE, &resolvedCommit, &cerr))) {
			return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
		}
		err := processOneCheckout(crepo, resolvedCommit, checkoutOpts.Subpath, destination, cancellable)
		if err != nil {
			return err
		}
	}
	return nil
}

func processOneCheckout(crepo *C.OstreeRepo, resolvedCommit *C.char, subpath, destination string, cancellable *glib.GCancellable) error {
	cdest := C.CString(destination)
	defer C.free(unsafe.Pointer(cdest))
	var gerr = glib.NewGError()
	cerr := (*C.GError)(gerr.Ptr())
	defer C.free(unsafe.Pointer(cerr))
	var repoCheckoutOptions C.OstreeRepoCheckoutOptions

	if checkoutOpts.UserMode {
		repoCheckoutOptions.mode = C.OSTREE_REPO_CHECKOUT_MODE_USER
	}
	if checkoutOpts.Union {
		repoCheckoutOptions.overwrite_mode = C.OSTREE_REPO_CHECKOUT_OVERWRITE_UNION_FILES
	}

	checkedOut := glib.GoBool(glib.GBoolean(C.ostree_repo_checkout_tree_at(crepo, &repoCheckoutOptions, C._at_fdcwd(), cdest, resolvedCommit, nil, &cerr)))
	if !checkedOut {
		return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
	}

	return nil
}

func processManyCheckouts(crepo *C.OstreeRepo, target string, cancellable *glib.GCancellable) error {
	return nil
}
