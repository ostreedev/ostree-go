package otbuiltin

import (
  "time"
  "errors"
  "strings"
  "strconv"
  "bytes"
  "unsafe"

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

func NewCommitOptions() commitOptions {
  var co commitOptions
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
func Commit(repoPath, commitPath string, opts commitOptions) (string, error) {
  options = opts


  var gerr = glib.NewGError()
  var cerr = (*C.GError)(gerr.Ptr())
  var metadata *glib.GVariant = nil
  var detachedMetadata *glib.GVariant = nil
  var flags C.OstreeRepoCommitModifierFlags = 0
  var modifier *C.OstreeRepoCommitModifier
  var modeAdds *glib.GHashTable
  var skipList *glib.GHashTable
  var mtree *C.OstreeMutableTree
  var root *C.GFile
  var objectToCommit *glib.GFile
  var skipCommit bool = false
  var ret string
  var commitChecksum string
  var stats C.OstreeRepoTransactionStats
  var filter_data C.CommitFilterData
  var cancellable *C.GCancellable

  cpath := C.CString(commitPath)
  csubject := C.CString(options.Subject)
  cbody := C.CString(options.Body)
  cbranch := C.CString(options.Branch)
  cparent := C.CString(options.Parent)

  // Open Repo function causes as Segfault.  Either openRepo or repo.native() has something wrong with it
  _, err := openRepo(repoPath)
  if err != nil {
    return "", err
  }
  // Create a repo struct from the path
  repoPathc := C.g_file_new_for_path(C.CString(repoPath))
  defer C.g_object_unref(repoPathc)
  crepo := C.ostree_repo_new(repoPathc)
  if !glib.GoBool(glib.GBoolean(C.ostree_repo_open(crepo, cancellable, &cerr))) {
    goto out
  }


  if !glib.GoBool(glib.GBoolean(C.ostree_repo_is_writable(crepo, &cerr))) {
    goto out
  }

  // If the user provided a stat override file
  if strings.Compare(options.StatOverrideFile, "") != 0 {
    modeAdds = glib.ToGHashTable(unsafe.Pointer(C._g_hash_table_new_full()))
    if err = parseFileByLine(options.StatOverrideFile, handleStatOverrideLine, modeAdds, cancellable); err != nil {
      goto out
    }
  }

  // If the user provided a skilist file
  if strings.Compare(options.SkipListFile, "") != 0 {
    skipList = glib.ToGHashTable(unsafe.Pointer(C._g_hash_table_new_full()))
    if err = parseFileByLine(options.SkipListFile, handleSkipListline, skipList, cancellable); err != nil {
      goto out
    }
  }

  if options.AddMetadataString != nil {
    err := parseKeyValueStrings(options.AddMetadataString, metadata)
    if err != nil {
      goto out
    }
  }

  if options.AddDetachedMetadataString != nil {
    err := parseKeyValueStrings(options.AddDetachedMetadataString, detachedMetadata)
    if err != nil {
      goto out
    }
  }

  if strings.Compare(options.Branch, "") == 0 {
    err = errors.New("A branch must be specified with --branch or use --orphan")
    goto out
  }

  if options.NoXattrs {
    C._ostree_repo_append_modifier_flags(&flags, C.OSTREE_REPO_COMMIT_MODIFIER_FLAGS_SKIP_XATTRS)
  }
  if options.GenerateSizes {
    C._ostree_repo_append_modifier_flags(&flags, C.OSTREE_REPO_COMMIT_MODIFIER_FLAGS_GENERATE_SIZES)
  }
  if !options.Fsync {
    C.ostree_repo_set_disable_fsync (crepo, C.TRUE)
  }

  if flags != 0 || options.OwnerUID >= 0 || options.OwnerGID >= 0 || strings.Compare(options.StatOverrideFile, "") != 0 || options.NoXattrs {
    filter_data.mode_adds = (*C.GHashTable)(modeAdds.Ptr())
    filter_data.skip_list = (*C.GHashTable)(skipList.Ptr())
    C._set_owner_uid ((C.guint32)(options.OwnerUID))
    C._set_owner_gid((C.guint32)(options.OwnerGID))
    modifier = C._ostree_repo_commit_modifier_new_wrapper(flags, &filter_data, nil)
  }

  if strings.Compare(options.Parent, "") != 0 {
    if strings.Compare(options.Parent, "none") == 0 {
      options.Parent = ""
    }
  } else if !options.Orphan {
    cerr = nil
    if !glib.GoBool(glib.GBoolean(C.ostree_repo_resolve_rev(crepo, cbranch, C.TRUE, &cparent, &cerr))) {
      goto out
    }
  }

  cerr = nil
  if !glib.GoBool(glib.GBoolean(C.ostree_repo_prepare_transaction(crepo, nil, cancellable, &cerr))) {
    goto out
  }

  cerr = nil
  if options.LinkCheckoutSpeedup && !glib.GoBool(glib.GBoolean(C.ostree_repo_scan_hardlinks(crepo, cancellable, &cerr))) {
    goto out
  }

  mtree = C.ostree_mutable_tree_new()

  if len(commitPath) == 0 && (len(options.Tree) == 0 || len(options.Tree[1]) == 0) {
    currentDir := (*C.char)(C.g_get_current_dir())
    objectToCommit = glib.ToGFile(unsafe.Pointer(C.g_file_new_for_path(currentDir)))
    C.g_free(currentDir)

    if !glib.GoBool(glib.GBoolean(C.ostree_repo_write_directory_to_mtree(crepo, (*C.GFile)(objectToCommit.Ptr()), mtree, modifier, cancellable, &cerr))) {
      goto out
    }
  } else if len(options.Tree) != 0 {
    var eq int = -1
    cerr = nil
    for tree := range options.Tree {
      eq = strings.Index(options.Tree[tree], "=")
      if eq == -1 {
        C._g_set_error_onearg(cerr, C.CString("Missing type in tree specification"), C.CString(options.Tree[tree]))
        goto out
      }
      treeType := options.Tree[tree][:eq]
      treeVal := options.Tree[tree][eq+1:]

      C._g_clear_object((**C.GObject)(objectToCommit.Ptr()))
      if strings.Compare(treeType, "dir") == 0 {
        objectToCommit = glib.ToGFile(unsafe.Pointer(C.g_file_new_for_path(C.CString(treeVal))))
        if !glib.GoBool(glib.GBoolean(C.ostree_repo_write_directory_to_mtree(crepo, (*C.GFile)(objectToCommit.Ptr()), mtree, modifier, cancellable, &cerr))) {
          goto out
        }
      } else if strings.Compare(treeType, "tar") == 0 {
        objectToCommit = glib.ToGFile(unsafe.Pointer(C.g_file_new_for_path(C.CString(treeVal))))
        if !glib.GoBool(glib.GBoolean(C.ostree_repo_write_directory_to_mtree(crepo, (*C.GFile)(objectToCommit.Ptr()), mtree, modifier, cancellable, &cerr))) {
          goto out
        }
      } else if strings.Compare(treeType, "ref") == 0 {
        if !glib.GoBool(glib.GBoolean(C.ostree_repo_read_commit(crepo, C.CString(treeVal), (**C.GFile)(objectToCommit.Ptr()), nil, cancellable, &cerr))) {
          goto out
        }

        if !glib.GoBool(glib.GBoolean(C.ostree_repo_write_directory_to_mtree(crepo, (*C.GFile)(objectToCommit.Ptr()), mtree, modifier, cancellable, &cerr))) {
          goto out
        }
      } else {
        C._g_set_error_onearg(cerr, C.CString("Missing type in tree specification"), C.CString(treeVal))
        goto out
      }
    }
  } else {
    objectToCommit = glib.ToGFile(unsafe.Pointer(C.g_file_new_for_path(cpath)))
    cerr = nil
    if !glib.GoBool(glib.GBoolean(C.ostree_repo_write_directory_to_mtree(crepo, (*C.GFile)(objectToCommit.Ptr()), mtree, modifier, cancellable, &cerr))) {
      goto out
    }
  }

  if modeAdds != nil && C.g_hash_table_size((*C.GHashTable)(modeAdds.Ptr())) > 0 {
    var hashIter *C.GHashTableIter

    var key, value C.gpointer

    C.g_hash_table_iter_init(hashIter, (*C.GHashTable)(modeAdds.Ptr()))

    for glib.GoBool(glib.GBoolean(C.g_hash_table_iter_next(hashIter, &key, &value))) {
      C._g_printerr_onearg(C.CString("Unmatched StatOverride path: "), C._gptr_to_str(key))
    }
    err = errors.New("Unmatched StatOverride paths")
    goto out
  }

  if skipList != nil && C.g_hash_table_size((*C.GHashTable)(skipList.Ptr())) > 0 {
    var hashIter *C.GHashTableIter
    var key, value C.gpointer

    C.g_hash_table_iter_init(hashIter, (*C.GHashTable)(skipList.Ptr()))

    for glib.GoBool(glib.GBoolean(C.g_hash_table_iter_next(hashIter, &key, &value))) {
      C._g_printerr_onearg(C.CString("Unmatched SkipList path: "), C._gptr_to_str(key))
    }
    err = errors.New("Unmatched SkipList paths")
    goto out
  }

  cerr = nil
  if !glib.GoBool(glib.GBoolean(C.ostree_repo_write_mtree(crepo, mtree, &root, cancellable, &cerr))) {
    goto out
  }

  if options.SkipIfUnchanged && strings.Compare(options.Parent, "") != 0 {
    var parentRoot *C.GFile

    cerr = nil
    if !glib.GoBool(glib.GBoolean(C.ostree_repo_read_commit(crepo, cparent, &parentRoot, nil, cancellable, &cerr))) {
      goto out
    }

    if glib.GoBool(glib.GBoolean(C.g_file_equal(root, parentRoot))) {
      skipCommit = true
    }
  }

  if !skipCommit {
    var updateSummary C.gboolean
    var timestamp C.guint64
    var ccommitChecksum = C.CString(commitChecksum)

    if options.Timestamp.IsZero() {
      var now *C.GDateTime = C.g_date_time_new_now_utc()
      timestamp = (C.guint64)(C.g_date_time_to_unix(now))
      C.g_date_time_unref(now)

      cerr = nil
      if !glib.GoBool(glib.GBoolean(C.ostree_repo_write_commit(crepo, cparent, csubject, cbody,
                     (*C.GVariant)(metadata.Ptr()), C._ostree_repo_file(root), &ccommitChecksum, cancellable, &cerr))) {
        goto out
      }
    } else {
      timestamp = (C.guint64)(options.Timestamp.Unix())
    }

    cerr = nil
    if !glib.GoBool(glib.GBoolean(C.ostree_repo_write_commit_with_time(crepo, cparent, csubject, cbody,
                   (*C.GVariant)(metadata.Ptr()), C._ostree_repo_file(root), timestamp, &ccommitChecksum, cancellable, &cerr))) {
      goto out
    }

    if detachedMetadata != nil {
      cerr = nil
      C.ostree_repo_write_commit_detached_metadata(crepo, ccommitChecksum, (*C.GVariant)(detachedMetadata.Ptr()), cancellable, &cerr)
    }

    if len(options.GpgSign) != 0 {
      for key := range options.GpgSign {
        cerr = nil
        if !glib.GoBool(glib.GBoolean(C.ostree_repo_sign_commit(crepo, (*C.gchar)(ccommitChecksum), (*C.gchar)(C.CString(options.GpgSign[key])), (*C.gchar)(C.CString(options.GpgHomedir)), cancellable, &cerr))) {
          goto out
        }
      }

      if options.Branch != "" {
        C.ostree_repo_transaction_set_ref(crepo, nil, cbranch, C.CString(commitChecksum))
      } else if !options.Orphan {
        err = errors.New("Error: commit must have a branch or be an orphan")
        goto out
      }

      cerr = nil
      if !glib.GoBool(glib.GBoolean(C.ostree_repo_commit_transaction(crepo, &stats, cancellable, &cerr))) {
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
      if glib.GoBool(glib.GBoolean(updateSummary)) &&  !glib.GoBool(glib.GBoolean(C.ostree_repo_regenerate_summary(crepo, nil, cancellable, &cerr))) {
        goto out
      }
    }
  } else {
    commitChecksum = options.Parent
  }

  if options.TableOutput {
    var buffer bytes.Buffer

    buffer.WriteString("Commit: ")
    buffer.WriteString(commitChecksum)
    buffer.WriteString("\nMetadata Total: ")
    buffer.WriteString(strconv.Itoa((int)(stats.metadata_objects_total)))
    buffer.WriteString("\nMetadata Written: ")
    buffer.WriteString(strconv.Itoa((int)(stats.metadata_objects_written)))
    buffer.WriteString("\nContent Total: ")
    buffer.WriteString(strconv.Itoa((int)(stats.content_objects_total)))
    buffer.WriteString("\nContent Written")
    buffer.WriteString(strconv.Itoa((int)(stats.content_objects_written)))
    buffer.WriteString("\nContent Bytes Written: ")
    buffer.WriteString(strconv.Itoa((int)(stats.content_bytes_written)))
    ret = buffer.String()
  } else {
    ret = commitChecksum
  }

  return ret, nil
  out:
    if crepo != nil { C.ostree_repo_abort_transaction(crepo, cancellable, nil) }
    if modifier != nil { C.ostree_repo_commit_modifier_unref(modifier) }
    if err != nil{
      return "", err
    }
    return "", glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
}

func parseKeyValueStrings(pairs []string, metadata *glib.GVariant) error {
  builder := C.g_variant_builder_new(C._g_variant_type(C.CString("a{sv}")))

  for iter := range pairs {
    index := strings.Index(pairs[iter], "=")
    if index >= 0 {
      var buffer bytes.Buffer
      buffer.WriteString("Missing '=' in KEY=VALUE metadata '%s'")
      buffer.WriteString(pairs[iter])
      return errors.New(buffer.String())
    }

    key := pairs[iter][:index]
    value := pairs[iter][index+1:]
    C._g_variant_builder_add_twoargs(builder, (*C.gchar)(C.CString("{sv}")), C.CString(key), C.CString(value))
  }

  metadata = glib.ToGVariant(unsafe.Pointer(C.g_variant_builder_end(builder)))
  C.g_variant_ref_sink((*C.GVariant)(metadata.Ptr()))

  return nil
}

func parseFileByLine(path string, fn handleLineFunc, table *glib.GHashTable, cancellable *C.GCancellable) error {
  var contents *C.char
  var file *glib.GFile
  var lines []string
  var gerr = glib.NewGError()
  cerr := (*C.GError)(gerr.Ptr())

  file = glib.ToGFile(unsafe.Pointer(C.g_file_new_for_path(C.CString(path))))
  if !glib.GoBool(glib.GBoolean(C.g_file_load_contents((*C.GFile)(file.Ptr()), cancellable, &contents, nil, nil, &cerr))) {
    return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
  }

  lines = strings.Split(C.GoString(contents), "\n")
  for line := range lines {
    if strings.Compare(lines[line], "") == 0 {
      continue
    }

    if err := fn(lines[line], table); err != nil {
      return glib.ConvertGError(glib.ToGError((unsafe.Pointer(cerr))))
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

  modeAdd = (C.guint)(C.g_ascii_strtod((*C.gchar)(C.CString(line)), nil))
  C.g_hash_table_insert((*C.GHashTable)(table.Ptr()), C.g_strdup((*C.gchar)(C.CString(line[space+1:]))), C._guint_to_pointer(modeAdd))

  return nil
}

func handleSkipListline(line string, table *glib.GHashTable) error {
  C.g_hash_table_add((*C.GHashTable)(table.Ptr()), C.g_strdup((*C.gchar)(C.CString(line))))

  return nil
}

/* func CommitFilter(self *C.OstreeRepo, path *C.char, fileInfo *C.GFileInfo, userData *C.CommitFilterData) C.OstreeRepoCommitFilterResult {
  var modeAdds *C.GHashTable
  var skipList *C.GHashTable
  var value C.gpointer

  if options.OwnerUID >= 0 {
    C.g_file_info_set_attribute_uint32(fileInfo, C.CString("unix::uid"), (C.guint32)(options.OwnerUID))
  }
  if options.OwnerGID >= 0 {
    C.g_file_info_set_attribute_uint32(fileInfo, C.CString("unix::gid"), (C.guint32)(options.OwnerGID))
  }

  if modeAdds != nil && glib.GoBool(glib.GBoolean(C.g_hash_table_lookup_extended(modeAdds, path, nil, &value))) {
    currentMode := C.g_file_info_get_attribute_uint32(fileInfo,  C.CString("unix::mode"))
    modeAdd := C._gpointer_to_uint(value)
    C.g_file_info_set_attribute_uint32(fileInfo, C.CString("unix::mode"), C._binary_or(currentMode, (C.guint32)(modeAdd)))
    C.g_hash_table_remove(modeAdds, path)
  }

  if skipList != nil && glib.GoBool(glib.GBoolean(C.g_hash_table_contains(skipList, path))) {
    C.g_hash_table_remove(skipList, path)
    return C.OSTREE_REPO_COMMIT_FILTER_SKIP
  }

  return C.OSTREE_REPO_COMMIT_FILTER_ALLOW
} */
