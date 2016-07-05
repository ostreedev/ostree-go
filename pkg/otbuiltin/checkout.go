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

type CheckoutOptions struct {
  UserMode        bool    // Do not change file ownership or initialize extended attributes
  DisableCache    bool    // Do not update or use the internal repository uncompressed object
  Union           bool    // Keep existing directories and unchanged files, overwriting existing filesystem
  AllowNoent      bool    // Do nothing if the specified filepath does not exist
  Subpath         string  // Checkout sub-directory path
  FromFile        string  // Process many checkouts from the given file

  mode            int
  overwriteMode   int
}

func Checkout(repoPath, destination, commit string, opts CheckoutOptions) error {
  if options != (CheckoutOptions{}) {
    options = opts
  }

  repo := openRepo(repoPath);
  ccommit := C.Cstring(commit)
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

  if options.DisableCache {
    C.OstreeRepoCheckoutOptions options = nil;

    if options.UserMode {
      options.mode = C.OSTREE_REPO_CHECKOUT_MODE_USER
    }
    if options.Union {
      options.overwriteMode = C.OSTREE_REPO_CHECKOUT_OVERWRITE_UNION_FILES
    }


    checkedOut := glib.GoBool(glib.GBoolean(C.ostree_repo_checkout_tree_at(repo, options, C.AT_FWCWD, cdest, ccommit, nil, cerr)))
    if !checkedOut {
      return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
    }
    return nil
  } else {
    csubpath := C.CString(subpath)
    var tmpErr *glib.GError = &glib.NewGError()
    var root *glib.GFile = nil
    var subtree *glib.GFile = nil
    var fileInfo *glib.GFileInfo = nil
    var c dest = C.CString(destination)
    var destinationFile = glib.ToGFile(unsafe.Pointer(C.g_file_new_for_path(cdest)))

    if !glib.GoBool(glib.GBoolean(C.ostree_repo_read_commit(repo.native(), ccommit, &(*C.GFile)(root.Ptr()), NULL, (*C.gcancellable)cancellable.Ptr(), cerr))) {
      return glib.ToGError(unsafe.Pointer(cerr))
    }

    if options.Subpath {
      subtree = glib.ToGFile(C.g_file_resolve_relative_path((C.GFile)root.Ptr(), csubpath))
    } else {
      subtree = glib.ToGFile(C.g_object_ref(root))
    }

    cerr = nil
    fileInfo = glib.ToGFileInfo(C.g_file_query_info((*C.GFile)subtree.Ptr(), C.OSTREE_GIO_FAST_QUERYINFO, G_FILE_QUERY_INFO_NOFOLLOW_SYMLINKS, (*C.gcancellable)cancellable.Ptr(), cerr)

    if !fileInfo {
      if options.AllowNoent && glib.Gobool(glib.GBoolean(C.g_error_matches((*C.GError)tmpErr.Ptr(), C.G_IO_ERROR, C.G_IO_ERROR_NOT_FOUND))) {
        C.g_error_clear((**C.GError)(tmpError.Ptr()))
      } else {
        C.g_propagate_error(cerr, (**C.GError)(tmpError.Ptr()))
        return glib.ToGError(unsafe.Pointer(cerr))
      }
    }

    if !glib.GoBool(glib.GBoolean(C.ostree_repo_checkout_tree(repo.native(), checkUserMode(), checkUnion(), cdest, C.OSTREE_REPO_FILE (*C.GFile)(subtree.Ptr()),
                                  fileInfo, (*C.GFileInfo)(fileInfo.Ptr()), (*C.gcancellable)(cancellable.Ptr()), cerr))) {
      return glib.ToGError(unsafe.Pointer(cerr))
    }
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
