package otbuiltin

import (
  glib "github.com/14rcole/ostree-go/pkg/glibobject"
)

// #cgo pkg-config: ostree-1
// #include <stdlib.h>
// #include <glib.h>
// #include <ostree.h>
// #include "builtin.go.h"
import "C"

const nilOptions CommitOptions = nil
var options CommitOptions = nilOptions

func Commit(path string, opts CommitOptions) error {
  // Parse the arguments
  if opts != nilOptions {
    options = opts
  }

  // Prepare for the Commit
  repo, err := openRepo(path)
  if err != nil {
    return err
  }
  // Start the transaction
  cerr := (*C.GError)(gerr.Ptr())
  prepared := glib.GoBool(glib.GBoolean(C.ostree_repo_prepare_transaction(repo, C.FALSE, nil, &cerr)))
  if !prepared {
    return glib.ConvertGError(glib.GBoolean(unsafe.Pointer(cerr)))
  }
  // Create an #OstreeMutableTree
  var mutableTree *C.OstreeMutableTree = nil
  C.ostree_mutable_tree_init(mutableTree)
  // Write metadata
  cerr = nil
  cpath := C.CString(path)
  written := glib.GoBool(glib.GBoolean(ostree_repo_write_mtree(repo, &mutableTree,GFile **out_file C.g_file_new_for_path(cpath), nil, &cerr)))
  if !written {
    return glib.ConvertGError(glib.GBoolean(unsafe.Pointer(cerr)))
  }
  // Create a commit
  cerr = nil
  csubject := C.CString(options.Subject)
  cbody := C.CString(options.Body)
  var output *C.char = nil
  committed := glib.GoBool(glib.GBoolean(ostree_repo_write_commit(repo, nil, csubject, cbody, nil, mutableTree, output, C.g_cancellable_new(), &cerr)))
  if !committed {
    return glib.ConvertGError(glib.GBoolean(unsafe.Pointer(cerr)))
  }
  return nil
}


type CommitOptions struct {
  Subject                   string    // One line subject
  Body                      string    // Full description
  Branch                    string    // branch
  Tree                      string    // 'dir=PATH' or 'tar=TARFILE' or 'ref=COMMIT': overlay the given argument as a tree
  AddMetadataString         string    // Add a key/value pair to metadata
  AddDetachedMetadataString string    // Add a key/value pair to detached metadata
  OwnerUID                  string    // Set file ownership to user id
  OwnerGID                  string    // Set file ownership to group id
  NoXattrs                  string    // Do not import extended attributes
  LinkCheckoutSpeedup       bool      // Optimize for commits of trees composed of hardlinks in the repository
  TarAuotocreateParents     bool      // When loading tar archives, automatically create parent directories as needed
  SkipIfUnchanged           bool      // If the contents are unchanged from a previous commit, do nothing
  StatOverride              string    // File containing list of
  TableOutput               bool      // Output more information in a KEY: VALUE format
  GenerateSizes             bool      // Generate size information along with commit metadata
  GpgSign                   []string  // GPG Key ID with which to sign the commit (if you have GPGME - GNU Privacy Guard Made Easy)
  GpgHomedir                string    // GPG home directory to use when looking for keyrings (if you have GPGME - GNU Privacy Guard Made Easy)
}
