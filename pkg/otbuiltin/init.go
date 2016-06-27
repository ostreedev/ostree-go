package otbuiltin

import (
       "errors"
       "strings"

       glib "github.com/14rcole/ostree-go/pkg/glibobject"
)

// #cgo pkg-config: ostree-1
// #include <stdlib.h>
// #include <glib.h>
// #include <ostree.h>
// #include "builtin.go.h"
import "C"

// Declare variables for options
var mode string = "bare"

func Init(path string, options map[string]string) (bool, error) {
  err := parseArgs(options)
  if err != nil {
    return false, err
  }

  //Create a repo struct from the path
  var gerr *glib.GError = nil
  var cerr = gerr.Native()
  cpath := C.CString(path)
  pathc := C.g_file_new_for_path(cpath)
  defer C.g_object_unref(pathc)
  crepo := C.ostree_repo_new(pathc)
  repo := repoFromNative(crepo)

  // If the repo exists in the filesystem, return an error but set exists to true
  var exists C.gboolean = 0
  success := glib.GoBool(C.ostree_repo_exists(crepo, &exists, &cerr))
  if !success {
    return false, glib.ConvertGError(cerr)
  } else if exists == 1{
    err = errors.New("repository already exists")
    return true, err
  }

  cerr = nil
  created := glib.GoBool(C.ostree_repo_create(crepo, C.OSTREE_REPO_MODE_BARE, nil, &cerr))
  if !created {
    return false, glib.ConvertGError(cerr)
  }
  return true, nil
}

func parseArgs (options map[string]string) error {
  for key, val := range options {
    if strings.EqualFold(key, "mode"){
      if strings.EqualFold("bare", val) {
        mode = "OSTREE_REPO_MODE_BARE"
      } else if strings.EqualFold("archive-z2", val) {
        mode = "OSTREE_REPO_MODE_ARCHIVE_Z2"
      } else {
        return errors.New("Invalid option for mode")
      }
    }
  }
  return nil
}
