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

type checkoutOptions struct {
  UserMode        bool    // Do not change file ownership or initialize extended attributes
  Union           bool    // Keep existing directories and unchanged files, overwriting existing filesystem
  AllowNoent      bool    // Do nothing if the specified filepath does not exist
  Subpath         string  // Checkout sub-directory path
  FromFile        string  // Process many checkouts from the given file

  mode            int
  overwriteMode   int
}

func NewCheckoutOptions() checkoutOptions {
  return checkoutOptions{}
}

func Checkout(repoPath, destination, commit string, opts checkoutOptions) error {
  if options != (CheckoutOptions{}) {
    options = opts
  }

  repo := openRepo(repoPath);
  ccommit := C.CString(commit)
  cdest := C.CString(destination)
  var gerr = glib.NewGError()
  cerr := (*C.GError)(gerr.Ptr())

  if options.FromFile {
    err := processManyCheckouts(repo, destination, (C.GCancellable)(cancellable.Ptr()))
    if err != nil {
      return err
    }
  } else {
    var resolvedCommit string
    if !glib.GoBool(glib.GBoolean(C.ostree_repo_resolve_rev(repo, ccommit, FALSE, &C.CString(resolvedCommit), cerr))) {
      return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
    }

    err := processOneCheckout(repo, resolvedCommit, options.Subpath, destination, (C.GCancellable)(cancellable.Ptr()))
    if err != nil {
      return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
    }
  }
  return nil
}

func processOneCheckout(OstreeRepo *repo, resolved_commit, subpath, destination string, cancellable glib.GCancellable) error {
  cdest := C.CString(destination)
  ccommit := C.CString(resolved_commit)
  var gerr = glib.NewGError()
  cerr := (*C.GError)(gerr.Ptr())
  var options C.OstreeRepoCheckoutOptions

  if options.UserMode {
    options.mode = C.OSTREE_REPO_CHECKOUT_MODE_USER
  }
  if options.Union {
    options.overwriteMode = C.OSTREE_REPO_CHECKOUT_OVERWRITE_UNION_FILES
  }


  checkedOut := glib.GoBool(glib.GBoolean(C.ostree_repo_checkout_tree_at(repo, options, C._at_fdcwd(), cdest, ccommit, nil, cerr)))
  if !checkedOut {
    return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
  }

  return nil
}

func processManyCheckouts(OstreeRepo *repo, target string, cancellable glib.GCancellable) error {
  return nil
}

func checkUserMode() int {
  if options.UserMode { return C.OSTREE_REPO_CHECKOUT_MODE_USER }
  return 0
}

func checkUnion() int {
  if options.Union { return C.OSTREE_REPO_CHECKOUT_OVERWRITE_UNION_FILES }
  return 0
}
