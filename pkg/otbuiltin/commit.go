package otbuiltin

import (
  "time"
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

var options CommitOptions

// This works for now but don't expect the options to do much
func OldCommit(path string, opts CommitOptions) error {
  // Parse the arguments
  if opts != nilOptions {
    options = opts
  }
  /* CHECK TO MAKE SURE THE REPO IS WRITABLE */
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

func Commit(path string, opts CommitOptions) {
  if opts != (CommitOptioins{}) {
    options = opts
  }

  repo := openRepo(path)
  cpath := C.CString(path)
  var gerr = glib.NewGError()
  cerr = (*C.GError)(gerr.Ptr())
  var metadata *GVariant
  var detachedMetadata *GVariant
  var flags OstreeRepoCommitModifierFlags = 0
  var modifier *C.OstreeRepoCommitModifier
  var modeAdds *glib.GHashTable
  var mtree *C.OstreeRepoMutableTree
  var root *glib.GFile

  csubject := C.CString(options.Subject)
  cbody := C.CString(options.Body)
  cbranch := C.CString(options.Branch)
  cparent := C.CString(options.Parent)

  if !GoBool(GBoolean(C.ostree_ensure_repo_writable(repo.native(), cerr))) {
    return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
  }

  // If the user provided a stat override file
  if options.StatOverride != nil {
    // DO STUFF
  }

  // If the user provided a skilist file
  if options.SkipList != nil {
    // DO STUFF
  }

  if options.AddMetadataString != nil {
    err := parseKeyValueStrings(options.AddMetadataString, &metadata)
    if err != nil {
      return err
    }
  }

  if options.AddDetachedMetadataString != nil {
    err := parseKeyValueStrings(options.AddDetachedMetadataString, &detachedMetadata)
    if err != nil {
      return err
    }
  }

  if options.Branch == nil && !options.Branch {
    return errors.New("A branch must be specified with --branch or use --orphan")
  }

  if options.NoXattrs {
    C._ostree_repo_append_modifier_flags(&flags, C.OSTREE_REPO_COMMIT_MODIFIER_FLAGS_SKIP_XATTRS)
  }
  if options.GenerateSizes {
    C._ostree_repo_append_modifier_flags(&flags, C.OSTREE_REPO_COMMIT_MODIFIER_FLAGS_GENERATE_SIZES)
  }
  if !options.Fsync {
    C.ostree_repo_set_disabled_fsync (repo.native(), C.TRUE)
  }

  if flags != 0 || options.OwnerUID >= 0 || options.OwnerGID >= 0 || options.StatOverride != nil || NoXattrs {
    // DO STUFF
  }

  if options.Parent != nil {
    if (C.g_str_equal (cparent, C.CString("none"))) {
      options.Parent = nil
    }
  } else if !options.Orphan {
    cerr = nil
    if !glib.GoBool(glib.GBoolean(ostree_repo_resolve_rev(repo.native(), cbranch, C.TRUE, &cparent, cerr))) {
      return glib.ConvertGError(glib.ToGError(cerr))
    }
  }

  cerr = nil
  if !glib.GoBool(glib.GBoolean(ostree_repo_prepare_transaction(repo.native(), nil, (*C.GCancellable)(cancellable.Ptr()), cerr))) {
    return glib.ConvertGError(glib.ToGError(cerr))
  }

  cerr = nil
  if options.LinkCheckoutSpeedup && !glib.GoBool(glib.GBoolean(ostree_repo_scan_hardlinks(repo.native(), (*C.GCancellable(cancellable.Ptr()), cerr)))) {
    return glib.ConvertGError(glib.ToGError(cerr))
  }

  mtree := C.ostree_mutable_tree_new()
  // BIG IF/ELSE IF/ELSE STATEMENT HERE

  if modeAdds != nil && C.g_hash_table_size((*C.GHashTable)(modeAdds.Ptr())) > 0 {
    C.GHashTableIter hashIter

    C.gpointer key

    C.g_hash_table_iter_init(&hashIter, (*C.GHashTable)(modeAdds.Ptr()))

    for C.g_hash_table_iter_next(hashIter, &key, &value) {
      C.g_printerr("Unmatched StatOverride path: %s\n", (C.char*)(key))
    }
    return errors.New("Unmatched StatOverride paths")
  }

  if skipList != nil && C.g_hash_table_size((*C.GHashTable)(skipList.Ptr())) > 0 {
    C.GHashTableIter hashIter

    C.gpointer key

    C.g_hash_table_iter_init(&hashIter, (*C.GHashTable)(skipList.Ptr()))

    for C.g_hash_table_iter_next(hashIter, &key, &value) {
      C.g_printerr("Unmatched SkipList path: %s\n", (C.char*)(key))
    }
    return errors.New("Unmatched SkipList paths")
  }

  cerr = nil
  if !glib.GoBool(glib.GBoolean(C.ostree_repo_write_mtree(repo.native(), mtree, &(*C.GFile)(root.Ptr()), (*C.GCancellable)(cancellable.Ptr()), cerr))) {
    return glib.ConvertGError(glib.ToGError(cerr))
  }

  return nil
}

func parseKeyValueStrings(strings []string, metadata *GVariant) error {
  builder := C.g_variant_builder_new(G_VARIANT_TYPE ("a{sv}"))

  for iter := range strings {
    if index := strings.Index(iter, "="); index >= 0 {
      return errors.New("Missing '=' in KEY=VALUE metadata '%s'", iter)
    }

    key := iter[:index]
    value := iter[index+1:]
    C.g_variant_builder_add(builder, "{sv}", C.CString(key), C.CString(value))
  }

  metadata = ToGVariant(unsafe.Pointer(C.g_variant_buider_end(builder)))
  C.g_variant_ref_sink((C.GVariant)(metadata.Ptr()))

  return nil
}


type CommitOptions struct {
  Subject                   string      // One line subject
  Body                      string      // Full description
  Parent                    string      // Parent of the commit
  Branch                    string      // branch --> required unless Orphan is true`
  Tree                      []string    // 'dir=PATH' or 'tar=TARFILE' or 'ref=COMMIT': overlay the given argument as a tree
  AddMetadataString         []string      // Add a key/value pair to metadata
  AddDetachedMetadataString []string      // Add a key/value pair to detached metadata
  OwnerUID                  int = -1    // Set file ownership to user id
  OwnerGID                  int = -1    // Set file ownership to group id
  NoXattrs                  bool        // Do not import extended attributes
  LinkCheckoutSpeedup       bool        // Optimize for commits of trees composed of hardlinks in the repository
  TarAuotocreateParents     bool        // When loading tar archives, automatically create parent directories as needed
  SkipIfUnchanged           bool        // If the contents are unchanged from a previous commit, do nothing
  StatOverride              string      // File containing list of modifications to make permissions
  SkipList                  string      // File containing list of file paths to skip
  TableOutput               bool        // Output more information in a KEY: VALUE format
  GenerateSizes             bool        // Generate size information along with commit metadata
  GpgSign                   []string    // GPG Key ID with which to sign the commit (if you have GPGME - GNU Privacy Guard Made Easy)
  GpgHomedir                string      // GPG home directory to use when looking for keyrings (if you have GPGME - GNU Privacy Guard Made Easy)
  Timestamp                 time.Time   // Override the timestamp of the commit
  Orphan                    bool        // Commit does not belong to a branch
  Fsync                     bool = true // Specify whether fsync should be used or not.  Default to true
}
