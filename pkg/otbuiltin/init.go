package otbuiltin

import (
       "errors"
       // #include "builtin.go.h"
       "strings"

       glib "github.com/14rcole/ostree-go/pkg/glibobject"
)

// #cgo pkg-config: ostree-1
// #include <stdlib.h>
// #include <glib.h>
// #include <ostree.h>
import "C"

// Declare variables for options
var mode string = "bare"

func Init(path string, options map[string]string) (bool, error) {
  err := parseArgs(options)
  if err != nil {
    return false, err
  }

  //Create a repo struct from the path
  var cerr *glib.GError = nil
  //var cerr *C.GError = nil
  cpath := C.CString(path)
  pathc := C.g_file_new_for_path(cpath)
  defer C.g_object_unref(pathc)
  crepo := C.ostree_repo_new(pathc)

  // If the repo exists in the filesystem, return an error but set exists to true
  var exists C.gboolean = 0
  success := glib.GoBool(glib.GBoolean(C.ostree_repo_exists(crepo, &exists, (**C.GError)(cerr.Raw()))))
  if exists == 1 {
    err = errors.New("repository already exists")
    return true, err
  } else if !success {
    return false, glib.ConvertGError(cerr)
  }

  cerr = nil
  gbool := C.ostree_repo_create(crepo, C.OSTREE_REPO_MODE_BARE, nil, (**C.GError)(cerr.Raw()))
  //gbool2 := (glib.CGBool)(gbool)
  created := glib.GoBool(glib.GBoolean(gbool))
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
