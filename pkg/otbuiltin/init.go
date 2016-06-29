package otbuiltin

import (
       "errors"
       "strings"
       "fmt"
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
var mode C.OstreeRepoMode = C.OSTREE_REPO_MODE_BARE

func Init(path string, options map[string]string) (bool, error) {
  err := parseArgs(options)
  if err != nil {
    return false, err
  }

  //Create a repo struct from the path
  var gerr = glib.NewGError()
  cerr := (*C.GError)(gerr.Ptr())
  cpath := C.CString(path)
  pathc := C.g_file_new_for_path(cpath)
  defer C.g_object_unref(pathc)
  crepo := C.ostree_repo_new(pathc)

  // If the repo exists in the filesystem, return an error but set exists to true
  var exists  C.gboolean = 0
  success := glib.GoBool(glib.GBoolean(C.ostree_repo_exists(crepo, &exists, &cerr)))
  if exists != 0 {
    err = errors.New("repository already exists")
    return true, err
  } else if !success {
    return false, glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
  }

  cerr = nil
  created := glib.GoBool(glib.GBoolean(C.ostree_repo_create(crepo, mode, nil, &cerr)))
  if !created {
    fmt.Println("Error is here")
    return false, glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
  }
  return true, nil
}

func parseArgs (options map[string]string) error {
  for key, val := range options {
    if strings.EqualFold(key, "mode"){
      if strings.EqualFold("bare", val) {
        mode = C.OSTREE_REPO_MODE_BARE
      } else if strings.EqualFold("archive-z2", val) {
        mode = C.OSTREE_REPO_MODE_ARCHIVE_Z2
      } else if strings.EqualFold("bare-user", val) {
        mode = C.OSTREE_REPO_MODE_BARE_USER
      } else {
        return errors.New("Invalid option for mode")
      }
    }
  }
  return nil
}
