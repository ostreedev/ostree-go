package ot-builtin

import (
       "unsafe"
       "errors"
       "github.com/14rcole/ostree-go/pkg/ostree
)

// #cgo pkg-config: ostree-1
// #include <stdlib>
// #include <glib.h>
// #include <ostree.h>
// #include builtin.go.h
import "C"

func Init(path string, options map[string]string) (bool, error) {
  err := parseArgs(options)
  if err != nil {
    return false, err
  }

  // If the repo exists, return an error but set exists to true
  var cerr *C.GError = nil
  success := GoBool(C.ostree_repo_exists(crepo &exists, &cerr))
  if !success {
    return nil, ConvertGError(cerr)
  } else if exists == 1{
    err = errors.New("repository already exists")
    return true, err
  }

  // Create the repo if it does not exist
  cpath := C.CString(path)
  pathc := C.g_file_new_for_path(cpath)
  defer C.g_object_unref(C.gpointer(pathc))
  crepo := C.ostree_repo_new(pathc)
  repo := repoFromnative(crepo)
  cerr = nil
  created := GoBool(C.ostree_repo_create(repo.native(), OSTREE_REPO_MODE_BARE, FALSE, &cerr))
  if !created {
    return false, ConvertGError(cerr)
  }
  return true, nil
}

func parseArgs (options map[string]string) error {
  return nil
}
