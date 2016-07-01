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

var options CheckoutOptions
type options struct {
  userOpts          CheckoutOptions
  mode              int
  overwriteMode     int
}
var opts options

type CheckoutOptions struct {
  UserMode        bool    // Do not change file ownership or initialize extended attributes
  DisableCache    bool    // Do not update or use the internal repository uncompressed object
  Union           bool    // Keep existing directories and unchanged files, overwriting existing filesystem
  AllowNoent      bool    // Do nothing if the specified filepath does not exist
  Subpath         string  // Checkout sub-directory path
  FromFile        string  // Process many checkouts from the given file
}

func Checkout(repo, filepath, commit string, options CheckoutOptions) error {
  if options != (CheckoutOptions{}) {
    opts.userOpts = options
  }
  return nil
}

func processOneCheckout(OstreeRepo *repo, resolved_commit, subpath, destination string, cancellable glib.GCancellable) error {
  cdest := C.CString(destination)
  ccommit := C.CString(resolved_commit)
  var gerr = glib.NewGError()
  cerr := (*C.GError)(gerr.Ptr())

  if opts.userOpts.DisableCache {
    C.OstreeRepoCheckoutOptions options = nil;

    if opts.userOpts.UserMode {
      opts.mode = C.OSTREE_REPO_CHECKOUT_MODE_USER
    }
    if opts.userOpts.Union {
      opts.overwriteMode = C.OSTREE_REPO_CHECKOUT_OVERWRITE_UNION_FILES
    }


    checkedOut := glib.GoBool(glib.GBoolean(C.ostree_repo_checkout_tree_at(repo, options, C.AT_FWCWD, cdest, ccommit, nil, cerr)))
    if !checkedOut {
      return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
    }
    return nil
  } else {
    csubpath := C.CString(subpath)
    var tmpErr glib.NewGError()
    var root *glib.GFile = nil
    var subtree *glib.GFile = nil
    var fileInfo *glib.GFileInfo = nil
    var dest = C.CString(destination)
    var destinationFile = glib.ToGFile(unsafe.Pointer(C.g_file_new_for_path(cdest)))

    if !glib.GoBool(glib.GBoolean(C.ostree_repo_read_commit(repo, ccommit, &(*C.GFile)(root.Ptr()), NULL, (*C.gcancellable)cancellable.Ptr(), cerr))) {
      return glib.ToGError(unsafe.Pointer(cerr))
    }

    if opts.userOpts.Subpath {
      subtree = glib.ToGFile(C.g_file_resolve_relative_path((C.GFile)root.Ptr(), csubpath))
    } else {
      subtree = glib.ToGFile(C.g_object_ref(root))
    }

    cerr = nil
    fileInfo = glib.ToGFileInfo(C.g_file_query_info((*C.GFile)subtree.Ptr(), C.OSTREE_GIO_FAST_QUERYINFO, G_FILE_QUERY_INFO_NOFOLLOW_SYMLINKS, (*C.gcancellable)cancellable.Ptr(), cerr)

    
  }
}

func processManyCheckouts(OstreeRepo *repo, target string, cancellable glib.GCancellable) error {

}
