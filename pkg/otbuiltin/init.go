package otbuiltin

import (
	"errors"
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

// Declare variables for options
var initOpts initOptions

type initOptions struct {
	Mode string // either bare, archive-z2, or bare-user

	repoMode C.OstreeRepoMode
}

func NewInitOptions() initOptions {
	io := initOptions{}
	io.Mode = "bare"
	io.repoMode = C.OSTREE_REPO_MODE_BARE
	return io
}

func Init(path string, options initOptions) (bool, error) {
	initOpts = options
	err := parseMode()
	if err != nil {
		return false, err
	}

	// Create a repo struct from the path
	var cerr *C.GError
	defer C.free(unsafe.Pointer(cerr))
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	pathc := C.g_file_new_for_path(cpath)
	defer C.g_object_unref(pathc)
	crepo := C.ostree_repo_new(pathc)

	// If the repo exists in the filesystem, return an error but set exists to true
	/* var exists C.gboolean = 0
	success := glib.GoBool(glib.GBoolean(C.ostree_repo_exists(crepo, &exists, &cerr)))
	if exists != 0 {
	  err = errors.New("repository already exists")
	  return true, err
	} else if !success {
	  return false, glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
	}*/

	cerr = nil
	created := glib.GoBool(glib.GBoolean(C.ostree_repo_create(crepo, initOpts.repoMode, nil, &cerr)))
	if !created {
		errString := glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr))).Error()
		if strings.Contains(errString, "File exists") {
			return true, glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
		}
		return false, glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
	}
	return true, nil
}

func parseMode() error {
  if strings.EqualFold(initOpts.Mode, "bare") {
    initOpts.repoMode = C.OSTREE_REPO_MODE_BARE
  } else if strings.EqualFold(initOpts.Mode, "bare-user") {
    initOpts.repoMode = C.OSTREE_REPO_MODE_BARE_USER
  } else if strings.EqualFold(initOpts.Mode, "archive-z2") {
    initOpts.repoMode = C.OSTREE_REPO_MODE_ARCHIVE_Z2
  } else {
    return errors.New("Invalid option for mode")
  }
  return nil
}
