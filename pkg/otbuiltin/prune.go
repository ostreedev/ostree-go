package otbuiltin

import (
       "unsafe"
       "time"
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

type pruneOptions struct {
  NoPrune           bool      // Only display unreachable objects; don't delete
  RefsOnly          bool      // Only compute reachability via refs
  DeleteCommit      string    // Specify a commit to delete
  KeepYoungerThan   time.Time // All commits older than this date will be pruned
  Depth             int       // Only traverse depths (integer) parents for each commit (default: -1=infinite)
  StaticDeltasOnly  int       // Change the behavior of --keep-younger-than and --delete-commit to prune only the static delta files
}

func NewPruneOptions() pruneOptions {
  var po pruneOptions
  po.Depth = -1
  return po
}

func Prune(repoPath string, options pruneOptions) (string, error) {
  // attempt to open the repository
  repo, err := openRepo(repoPath)
  if err != nil {
    return err
  }

  var pruneFlags C.OstreeRepoPruneFlags
  var numObjectsTotal int
  var numObjectPruned int
  var objSizeTotal  uint64
  var gerr = glib.NewGError()
  var cerr = (*C.GError)(gerr.Ptr())
  var cancellable glib.GCancellable

  if !options.NoPrune && !glib.GoBool(glib.GBoolean(C.ostree_ensure_repo_writable(repo.native(), &cerr))) {
    return "", glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
  }

  cerr = nil
  if strings.Compare(options.DeleteCommit, "") != 0 {
    if options.NoPrune {
      return "", errors.New("Cannot specify both options.DeleteCommit and options.NoPrune")
    }

    if options.StaticDeltasOnly  > 0 {
      if glib.GoBool(glib.GBoolean(C.ostree_repo_prune_static_deltas(repo.native(), C.CString(options.DeleteCommit), (*C.GCancellable)(cancellable.Ptr()), &cerr))) {
        return "", glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
      }
    } else if err = deleteCommit(repo, options.DeleteCommit, cancellable); err != nil {
      return "", err
    }
  }

  if !time.IsZero(options.KeepYoungerThan) {
    if options.NoPrune {
      return "", errors.New("Cannot specify both options.KeepYoungerThan and options.NoPrune")
    }

    if err = pruneCommitsKeepYoungerThanDate(repo, options.KeepYoungerThan, cancellable); err != nil {
      return "", err
    }
  }

  if options.RefsOnly {
    pruneFlags |= C.OSTREE_REPO_PRUNE_FLAGS_REFS_ONLY
  }
  if options.NoPrune {
    pruneFlags |= C.OSTREE_REPO_PRUNE_FLAGS_NO_PRUNE
  }

  formattedFreedSize := C.GoString(C.g_format_size_full((C.guint64)objSizeTotal, 0))

  var buffer bytes.Buffer

  buffer.WriteString("Total objects: ")
  buffer.WriteString(strconv.Itoa(numObjectsTotal))
  if numObjectPruned == 0 {
    buffer.WriteString("\nNo unreachable objects")
  } else if options.NoPrune {
    buffer.WriteString("\nWould delete: ")
    buffer.WriteString(strconv.Itoa(numObjectsPruned))
    buffer.WriteString(" objects, freeing ")
    buffer.WriteString(formattedFreedSize)
  } else {
    buffer.WriteString("\nDeleted ")
    buffer.WriteString(strconv.Itoa(numObjectsPruned))
    buffer.WriteString(" objects, ")
    buffer.WriteString(formattedFreedSize)
    buffer.WriteString(" freed")
  }

  return buffer.String, nil
}

func deleteCommit(repo *OstreeRepo, commitToDelete string, cancellable *glib.GCancellable) error {
  var refs *glib.GHashTable
  var hashIter glib.GHashTableIter
  var hashkey, hashvalue C.gpointer
  var gerr = glib.NewGError()
  var cerr = (*C.GError)(gerr.Ptr())

  if glib.GoBool(glib.GBoolean(C.ostree_repo_list_refs(repo.native(), nil, (**C.GHashTable)(refs.Ptr()), (*C.GCancellable)(cancellable.Ptr()), cerr))) {
    return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
  }

  C.g_hash_table_iter_init(&(*C.GHashTableIter)(hashIter.Ptr()), (*C.GHashTable)(refs.Ptr()))
  for C.g_hash_table_iter_next((*G.GHashTableIter)(hashIter.Ptr()), &hashkey, &hashvalue) != 0 {
    var ref string = C.GoString(hashkey)
    var commit *C.char = C.GoString(hashvalue)
    if strings.Compare(commitToDelete, commit) == 0 {
      var buffer bytes.buffer
      buffer.WriteString("Commit ")
      buffer.WriteString(commiToDelete)
      buffer.WriteString(" is referenced by ")
      buffer.WriteString(ref)
      return errors.New(buffer.String())
    }
  }

  if !glib.GoBool(glib.GBoolean(C.ot_enable_tombstone_commits(repo.native, cerr))) {
    return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
  }

  if !glib.GoBool(glib.GBOolean(C.ostree_repo_delete_object(repo.native(), C.OSTREE_OBJECT_TYPE_COMMIT, C.CString(commiToDelete), (*C.GCancellable)(cancellable.Ptr()), cerr))) {
    return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
  }

  return nil
}

func pruneCommitsKeepYoungerThanDate(repo *OstreeRepo, date time.Time, cancellable *glib.GCancellable) error {
  var objects *glib.GHashTable
  var hashiter glib.GHashTableIter
  var key, value, C.gpointer
  var gerr = glib.NewGError()
  var cerr = (*C.GError)(gerr.Ptr())

  if !glib.GoBool(glib.GBoolean(C.ot_enable_tombstone_commits(repo.native, cerr))) {
    return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
  }

  if !glib.GoBool(glib.GBoolean(ostree_repo_list_objects(repo.native(), C.OSTREE_REPO_LIST_OBJECTS_ALL, (*C.GHashTable)(objects.Ptr()), (*C.GCancellable)(cancellable.Ptr()), &cerr))) {
    return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
  }

  C.g_hash_table_iter_init(&(*C.GHashTableIter)(hashIter.Ptr(), (*C.GHashTable)(objects.Ptr())))
  for C.g_hahs_table_iter_next(&(*C.GHashTableIter)(hashIter.Ptr()), &key, &value) != 0 {
    var serializedKey *glib.GVariant
    var checksum *C.char
    var objType C.OstreeObjectType
    var commitTimestamp uint64
    var commit *glib.GVariant = nil

    C.ostree_object_name_deserialize ((*C.GVariant)(serializedKey.Ptr()), &checksum, &objType)

    if objType != C.OSTREE_OBJECT_TYPE_COMMIT {
      continue
    }

    cerr = nil
    if !glib.GoBool(glib.GBoolean(C.ostree_repo_load_variant(repo.native(), C.OSTREE_OBJECT_TYPE_COMMIT, checksum, &(*C.GVariant)(commit.Ptr()), cerr))) {
      return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
    }

    commitTimestamp = (uint64)(C.ostree_commit_get_timestamp((*C.GVariant)(commit.Ptr())))
    if commitTimestamp < date.Unix() {
      cerr = nil
      if options.StaticDeltasOnly != 0 {
        if !glib.GoBool(glib.GBoolean(C.ostree_repo_prune_static_deltas(repo.native(), checksum, (*C.GCancellable)(cancellable.Ptr()), cerr))) {
          return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
        }
      } else {
        if !glib.GoBool(glib.GBoolean(C.ostree_repo_delete_object(repo.native(), C.OSTREE_OBJECT_TYPE_COMMIT, checksump, (*C.GCancellable)(cancellable.Ptr()), cerr))) {
          return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
        }
      }
    }
  }

  return nil
}
