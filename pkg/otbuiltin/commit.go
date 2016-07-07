package otbuiltin

import (
  "time"
  "errors"
  "strings"
  "strconv"
  "bytes"

  glib "github.com/14rcole/ostree-go/pkg/glibobject"
)

// #cgo pkg-config: ostree-1
// #include <stdlib.h>
// #include <glib.h>
// #include <ostree.h>
// #include "builtin.go.h"
import "C"

var options commitOptions

type handleLineFunc func(string, *glib.GHashTable) error

type commitOptions struct {
  Subject                   string          // One line subject
  Body                      string          // Full description
  Parent                    string          // Parent of the commit
  Branch                    string          // branch --> required unless Orphan is true`
  Tree                      []string        // 'dir=PATH' or 'tar=TARFILE' or 'ref=COMMIT': overlay the given argument as a tree
  AddMetadataString         []string        // Add a key/value pair to metadata
  AddDetachedMetadataString []string        // Add a key/value pair to detached metadata
  OwnerUID                  int             // Set file ownership to user id
  OwnerGID                  int             // Set file ownership to group id
  NoXattrs                  bool            // Do not import extended attributes
  LinkCheckoutSpeedup       bool            // Optimize for commits of trees composed of hardlinks in the repository
  TarAuotocreateParents     bool            // When loading tar archives, automatically create parent directories as needed
  SkipIfUnchanged           bool            // If the contents are unchanged from a previous commit, do nothing
  StatOverrideFile          string          // File containing list of modifications to make permissions
  SkipListFile              string          // File containing list of file paths to skip
  TableOutput               bool            // Output more information in a KEY: VALUE format
  GenerateSizes             bool            // Generate size information along with commit metadata
  GpgSign                   []string        // GPG Key ID with which to sign the commit (if you have GPGME - GNU Privacy Guard Made Easy)
  GpgHomedir                string          // GPG home directory to use when looking for keyrings (if you have GPGME - GNU Privacy Guard Made Easy)
  Timestamp                 time.Time       // Override the timestamp of the commit
  Orphan                    bool            // Commit does not belong to a branch
  Fsync                     bool            // Specify whether fsync should be used or not.  Default to true
}

func NewCommitOptions() {
  var co CommitOptions
  co.OwnerUID = -1
  co.OwnerGID = -1
  co.Fsync = true
  return co
}
/*
// This works for now but don't expect the options to do much
func OldCommit(path string, opts CommitOptions) error {
  // Parse the arguments
  if opts != nilOptions {
    options = opts
  }
  /* CHECK TO MAKE SURE THE REPO IS WRITABLE
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
*/
func Commit(repoPath, commitPath string, opts CommitOptions) error {
  if opts != (CommitOptioins{}) {
    options = opts
  }

  repo := openRepo(repo)
  cpath := C.CString(commitPath)
  var gerr = glib.NewGError()
  cerr = (*C.GError)(gerr.Ptr())
  var metadata *GVariant = nil
  var detachedMetadata *GVariant = nil
  var flags C.OstreeRepoCommitModifierFlags = 0
  var modifier *C.OstreeRepoCommitModifier
  var modeAdds *glib.GHashTable
  var skipList *glib.GHashTable
  var mtree *C.OstreeMutableTree
  var root *glib.GFile
  var objectToCommit *glib.GFile
  var skipCommit bool = false
  var ret string = nil
  var commitChecksum string
  var stats C.OstreeRepoTransactionStats
  var filter_data C.CommitFilterData = nil

  csubject := C.CString(options.Subject)
  cbody := C.CString(options.Body)
  cbranch := C.CString(options.Branch)
  cparent := C.CString(options.Parent)

  if !glib.GoBool(glib.GBoolean(C.ostree_ensure_repo_writable(repo.native(), cerr))) {
    return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
  }

  // If the user provided a stat override file
  if options.StatOverrideFile != nil {
    modeAdds = glib.ToGHashTable(unsafe.Pointer(C.g_hash_table_new_full(C.g_str_hash, C.g_str_equal, C.g_free, NULL)))
    if err = parseFileByLine(options.StatOverrideFile, handleStatOverrideLine, modeAdds, cancellable); err != nil {
      goto out
    }
  }

  // If the user provided a skilist file
  if options.SkipListFile != nil {
    skipList = glib.ToGHashTable(unsafe.Pointer(C.g_hash_table_new_full(C.g_str_hash, C.g_str_equal, C.g_free, NULL)))
    if err = parseFileByLine(options.SkipListFile, handleSkipListline, skipList, cancellable); err != nil {
      goto out
    }
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

  if flags != 0 || options.OwnerUID >= 0 || options.OwnerGID >= 0 || options.StatOverrideFile != nil || NoXattrs {
    filter_data.mode_adds = (*C.GHashTable)(modeAdds.Ptr())
    filter_data.skip_list = (*C.GHashTable)(skipList.Ptr())
    modifier = C.ostree_repo_commit_modifier_new(flags, C._commit_filter, &filter_data, nil)
  }

  if options.Parent != nil {
    if (C.g_str_equal (cparent, C.CString("none"))) {
      options.Parent = nil
    }
  } else if !options.Orphan {
    cerr = nil
    if !glib.GoBool(glib.GBoolean(ostree_repo_resolve_rev(repo.native(), cbranch, C.TRUE, &cparent, cerr))) {
      goto out
    }
  }

  cerr = nil
  if !glib.GoBool(glib.GBoolean(ostree_repo_prepare_transaction(repo.native(), nil, (*C.GCancellable)(cancellable.Ptr()), cerr))) {
    goto out
  }

  cerr = nil
  if options.LinkCheckoutSpeedup && !glib.GoBool(glib.GBoolean(ostree_repo_scan_hardlinks(repo.native(), (*C.GCancellable)(cancellable.Ptr()), cerr))) {
    goto out
  }

  mtree := C.ostree_mutable_tree_new()

  if len(commitPath) == 0 && (len(options.Tree) == 0 || len(options.Tree[1]) == 0) {
    currentDir := C.g_get_current_dir()
    objectToCommit = glib.ToGFile(unsafe.Pointer(C.g_file_new_for_path(currentDir)))
    C.g_free(currentDir)

    if !glib.GoBool(glib.GBoolean(ostree_repo_write_directory_to_mtree(repo.native(), (*C.GFile)(objectToCommit.Ptr()), mtree, modifier, (*C.GCancellable)(cancellable.Ptr()), cerr))) {
      goto out
    }
  } else if len(options.Tree) != 0 {
    var eq int = -1
    cerr = nil
    for tree := range options.Tree {
      eq = strings.Index(tree, "=")
      if eq == -1 {
        C.g_set_error(cerr, C.G_IO_ERROR, C.G_IO_ERROR_FAILED, "Missing type in tree specification %s", tree)
        goto out
      }
      treeType := tree[:eq]
      tree = tree[eq+1:]

      C.g_clear_object((*C.GFile)(objectToCommit.Ptr()))
      if strings.Compare(treeType, "dir") == 0 {
        objectToCommit = glib.ToGFile(C.g_file_new_for_path(C.CString(tree)))
        if !glib.GoBool(glib.GBoolean(ostree_repo_write_directory_to_mtree(repo.native(), (*C.GFile)(objectToCommit.Ptr()), mtree, modifier, (*C.GCancellable)(cancellable.Ptr()), cerr))) {
          goto out
        }
      } else if strings.Compare(treeType, "tar") == 0 {
        objectToCommit = glib.ToGFile(C.g_file_new_for_path(C.CString(tree)))
        if !glib.GoBool(glib.GBoolean(ostree_repo_write_directory_to_mtree(repo.native(), (*C.GFile)(objectToCommit.Ptr()), mtree, modifier, (*C.GCancellable)(cancellable.Ptr()), cerr))) {
          goto out
        }
      } else if strings.Compare(treeType, "ref") {
        if !glib.GoBool(glib.GBoolean(C.ostree_repo_read_commit(repo.native, C.CString(tree), &(*C.GFile)(objectToCommit.Ptr()), mtree, modifier, (*C.GCancellable)(cancellable.Ptr()), cerr))) {
          goto out
        }

        if !glib.GoBool(glib.GBoolean(ostree_repo_write_directory_to_mtree(repo.native(), (*C.GFile)(objectToCommit.Ptr()), mtree, modifier, (*C.GCancellable)(cancellable.Ptr()), cerr))) {
          goto out
        }
      } else {
        C.g_set_error(cerr, C.G_IO_ERROR, C.G_IO_ERROR_FAILED, "Missing type in tree specification %s", tree)
        goto out
      }
    }
  } else {
    objectToCommit = glib.ToGFile(unsafe.Pointer(C.g_file_new_for_path(cpath)))
    cerr = nil
    if !glib.GoBool(glib.GBoolean(ostree_repo_write_directory_to_mtree(repo.native(), (*C.GFile)(objectToCommit.Ptr()), mtree, modifier, (*C.GCancellable)(cancellable.Ptr()), cerr))) {
      goto out
    }
  }

  if modeAdds != nil && C.g_hash_table_size((*C.GHashTable)(modeAdds.Ptr())) > 0 {
    var hashIter C.GHashTableIter

    var key C.gpointer

    C.g_hash_table_iter_init(&hashIter, (*C.GHashTable)(modeAdds.Ptr()))

    for glib.GoBool(glib.GBoolean(C.g_hash_table_iter_next(hashIter, &key, &value))) {
      C.g_printerr("Unmatched StatOverride path: %s\n", C._gptr_to_str(key))
    }
    return errors.New("Unmatched StatOverride paths")
  }

  if skipList != nil && C.g_hash_table_size((*C.GHashTable)(skipList.Ptr())) > 0 {
    var hashIter C.GHashTableIter
    var key C.gpointer

    C.g_hash_table_iter_init(&hashIter, (*C.GHashTable)(skipList.Ptr()))

    for glib.GoBool(glib.GBoolean(C.g_hash_table_iter_next(hashIter, &key, &value))) {
      C.g_printerr("Unmatched SkipList path: %s\n", C._gptr_to_str(key))
    }
    return errors.New("Unmatched SkipList paths")
  }

  cerr = nil
  if !glib.GoBool(glib.GBoolean(C.ostree_repo_write_mtree(repo.native(), mtree, &(*C.GFile)(root.Ptr()), (*C.GCancellable)(cancellable.Ptr()), cerr))) {
    goto out
  }

  if options.SkipIfUnchanged && options.Parent != nil {
    var parentRoot *glib.GFile

    cerr = nil
    if !glib.GoBool(glib.GBoolean(C.ostree_repo_read_commit(repo.native(), cparent, (*C.GFile)(parentRoot.Ptr()), NULL, (*C.GCancellable)(cancellable.Ptr()), cerr))) {
      goto out
    }

    if glib.GoBool(glib.GBoolean(C.g_file_equal((*C.GFile)(root.Ptr()), (*C.GFile)(parentRoot.Ptr())))) {
      skipCommit = true
    }
  }

  if !skipCommit {
    var updateSummary C.gboolean
    var timestamp C.guint64
    if options.Timestamp.IsZero() {
      var now *C.GDateTime = g_date_time_new_now_utc()
      timestamp = C.g_date_time_to_unix(now)
      C.g_date_time_unref(now)

      cerr = nil
      if !glib.GoBool(glib.GBoolean(ostree_repo_write_commit(repo.native(), cparent, csubject, cbody,
                     (*C.GVariant)(metadata.Ptr()), C.OSTREE_REPO_FILE((*C.GFile)(root.Ptr())), &C.CString(commitChecksum), (*C.GCancellable)(cancellable.Ptr()), cerr))) {
        goto out
      }
    } else {
      var ts C.timespec
      var timestampStr = strconv.FormatInt(options.Timestamp.Unix(), 10)
      if !C.parse_datetime(&ts, timestampStr, nil) {
        C.g_set_error(cerr, C.G_IO_ERROR, C.G_IO_ERROR_FAILED, "Could not parse %s", timestampStr)
        goto out
      }
      timestamp =  ts.tv_sec
    }

    cerr = nil
    if !glib.GoBool(glib.GBoolean(ostree_repo_write_commit_with_time(repo.native(), cparent, csubject, cbody,
                   (*C.GVariant)(metadata.Ptr()), C.OSTREE_REPO_FILE((*C.GFile)(root.Ptr())), timestamp, &C.CString(commitChecksum), (*C.GCancellable)(cancellable.Ptr()), cerr))) {
      goto out
    }

    if detachedMetadata != nil {
      cerr = nil
      C.ostree_repo_write_commit_detached_metadata(repo.native(), C.CString(commitChecksum), (*C.GVariant)(detachedMetadata.Ptr()), (*C.GCancellable)(cancellable.Ptr()), cerr)
    }

    if len(options.GpgSign) != 0 {
      for key := range options.GpgSign {
        cerr = nil
        if !glib.GoBool(glib.GBoolean(ostree_repo_sign_commit(repo.native(), C.CString(commitChecksum), C.CString(key), C.CString(options.GpgHomedir), (*C.GCancellable)(cancellable.Ptr()), cerr))) {
          goto out
        }
      }

      if options.Branch != "" {
        C.ostree_repo_transaction_set_ref(repo.native(), nil, cbranch, C.CString(commitChecksum))
      } else if !options.Orphan {
        return errors.New("Error: commit must have a branch or be an orphan")
      }

      cerr = nil
      if !glib.GoBool(glib.GBoolean(ostree_repo_commit_transaction(repo.native(), &stats, (*C.GCancellable)(cancellable.Ptr()), cerr))) {
        goto out
      }

      /* The default for this option is FALSE, even for archive-z2 repos,
       * because ostree supports multiple processes committing to the same
       * repo (but different refs) concurrently, and in fact gnome-continuous
       * actually does this.  In that context it's best to update the summary
       * explicitly instead of automatically here. */
      /*
      TODO: I think this function is declared outside of libostree so I have to hunt it down
      This is the C code:

      if (!ot_keyfile_get_boolean_with_default (ostree_repo_get_config (repo), "core",
                                                "commit-update-summary", FALSE,
                                                &update_summary, error))
        goto out;
      */

      cerr = nil
      if glib.GoBool(updateSummary) &&  !glib.GoBool(glib.GBoolean(ostree_repo_regenerate_summary(repo.native(), nil, (*C.GCancellable)(cancellable), cerr))) {
        goto out
      }
    }
  } else {
    commitChecksum = parent
  }

  if options.TableOutput {
    var buffer bytes.Buffer

    buffer.WriteString("Commit: ")
    buffer.WriteString(commitChecksum)
    buffer.WriteString("\nMetadata Total: ")
    buffer.WriteString(stats.metadata_objects_total)
    buffer.WriteString("\nMetadata Written: ")
    buffer.WriteString(stats.metadata_objects_written)
    buffer.WriteString("\nContent Total: ")
    buffer.WriteString(stats.content_objects_total)
    buffer.WriteString("\nContent Written")
    buffer.WriteString(stats.content_objects_written)
    buffer.WriteString("\nContent Bytes Written: ")
    buffer.WriteString(stats.content_bytes_written)
    ret = buffer.String()
  } else {
    ret = commitChecksum
  }

  return ret
  out:
    if repo.isInitialized() { C.ostree_repo_abort_transaction(repo.native(), (*C.GCancellable)(cancellable.Ptr()), NULL) }
    if modifier != nil { C.ostree_repo_commit_modifier_unref(modifier) }
    return nil, glib.ToGError((*C.GError)(unsafe.Pointer(cerr)))
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

  metadata = ToGVariant(unsafe.Pointer(C.g_variant_builder_end(builder)))
  C.g_variant_ref_sink((C.GVariant)(metadata.Ptr()))

  return nil
}

func parseFileByLine(path string, fn handleLineFunc, table *glib.GHashTable, cancellable *glib.GCancellable) error {
  var contents C.CString
  var file *glib.GFile
  var lines []string
  var gerr = glib.NewGError()
  cerr = (*C.GError)(gerr.Ptr())

  file = glib.ToGFile(unsafe.Pointer(C.g_file_new_for_path(C.CString(path))))
  if !glib.GoBool(glib.GBoolean(C.g_file_load_contents((*C.GFile)(file.Ptr()), (*C.GCancellable)(cancellable.Ptr()), &contents, nil, nil, cerr))) {
    return glib.ToGError((*C.GError)(unsafe.Pointer(cerr)))
  }

  lines = strings.Split(C.GoString(contents))
  for line := range lines {
    if line == nil {
      continue
    }

    if err := handleLineFunc(line, table); err != nil {
      return glib.ToGError((*C.GError)(unsafe.Pointer(cerr)))
    }
  }
  return nil
}

func handleStatOverrideLine(line string, table *glib.GHashTable) error {
  var space int
  var modeAdd C.guint

  if space = strings.IndexRune(line, ' '); space == -1 {
    return errors.New("Malformed StatOverrideFile (no space found)")
  }

  modeAdd = (C.guint32)(C.gint32)(C.g_ascii_strtod(C.CString(line), nil))
  C.g_hash_table_insert((*C.GHashTable)(table.Ptr()), C.g_strdup(C.CString(line[space+1:]), C.GUINT_TO_POINTER((gint32)(modeAdd))))

  return nil
}

func handleSkipListline(line string, table *glib.GHashTable) error {
  C.g_hash_table_add((*C.GHashTable)(table.Ptr()), C.g_strdup(C.CString(line)))

  return nil
}

// export commitFilter
func commitFilter(self *C.OstreeRepo, path C.CString, fileInfo *C.GFileInfo, userData *C.CommitFilterData) C.OstreeRepoCommitFilterResult {
  var modeAdds *C.GHashTable = userData.modeAdds
  var skipList *C.GHashTable = userData.skipList
  var value C.gpointer

  if options.OwnerUID >= 0 {
    C.g_file_info_set_attribute_uint32(fileInfo, "unix::uid", options.OwnerUID)
  }
  if options.OwnerGID >= 0 {
    C.g_file_info_set_attribute_uint32(fileInfo, "unix::gid", options.OwnerGID)
  }

  if modeAdds != nil && glib.GoBool(glib.GBoolean(C.g_hash_table_lookup_extended(modeAdds, path, nil, &value))) {
    currentMode := C.g_file_info_get_attribute_uint32(fileInfo, "unix::mode")
    modeAdd := C.GPOINTER_TO_UINT(value)
    C.g_file_info_set_attribute_uint32(fileInfo, "unix::mode", currentMode | modeAdd)
    C.g_hash_table_remove(modeAdds, path)
  }

  if skipList != nil && glib.GoBool(glib.GBoolean(C.g_hash_table_contains(skipList, path))) {
    C.g_hash_table_remove(skipList, path)
    return C.OSTREE_REPO_COMMIT_FILTER_SKIP
  }

  return C.OSTREE_REPO_COMMIT_FILTER_ALLOW
}
