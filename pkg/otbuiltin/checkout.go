package otbuiltin

import (
       "unsafe"
       "strings"
       "fmt"

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
  checkoutOpts = opts

  repo, err := openRepo(repoPath);
  if err != nil {
    fmt.Println("error opening repo")
    return err
  }

  var cancellable *glib.GCancellable
  ccommit := C.CString(commit)
  var gerr = glib.NewGError()
  cerr := (*C.GError)(gerr.Ptr())

  if strings.Compare(checkoutOpts.FromFile, "") != 0 {
    err := processManyCheckouts(repo, destination, cancellable)
    if err != nil {
      return err
    }
  } else {
    var resolvedCommit string
    cresolvedCommit := C.CString(resolvedCommit)
    if !glib.GoBool(glib.GBoolean(C.ostree_repo_resolve_rev(repo.native(), ccommit, C.FALSE, &cresolvedCommit, &cerr))) {
      return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
    }

    err := processOneCheckout(repo, resolvedCommit, checkoutOpts.Subpath, destination, cancellable)
    if err != nil {
      return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
    }
  }
  return nil
}

func processOneCheckout(repo *Repo, resolved_commit, subpath, destination string, cancellable *glib.GCancellable) error {
  cdest := C.CString(destination)
  ccommit := C.CString(resolved_commit)
  var gerr = glib.NewGError()
  cerr := (*C.GError)(gerr.Ptr())
  var repoCheckoutOptions C.OstreeRepoCheckoutOptions

  if checkoutOpts.UserMode {
    checkoutOpts.mode = C.OSTREE_REPO_CHECKOUT_MODE_USER
  }
  if checkoutOpts.Union {
    checkoutOpts.overwriteMode = C.OSTREE_REPO_CHECKOUT_OVERWRITE_UNION_FILES
  }


  checkedOut := glib.GoBool(glib.GBoolean(C.ostree_repo_checkout_tree_at(repo.native(), &repoCheckoutOptions, C._at_fdcwd(), cdest, ccommit, nil, &cerr)))
  if !checkedOut {
    return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
  }

  return nil
}

func processManyCheckouts(repo *Repo, target string, cancellable *glib.GCancellable) error {
  return nil
}
