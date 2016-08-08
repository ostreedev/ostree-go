package otbuiltin

import (
  "unsafe"
  "fmt"
  "bytes"
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

var fsckOpts FsckOptions

type FsckOptions struct {
  Quiet   bool  // Only print error messages
  Delete  bool  // Remove corrupted objects
  AddTombstones   bool  // Add tombstone commit for referenced but missing commits
}

func NewFsckOptions() FsckOptions {
  return FsckOptions{}
}

func Fsck(repoPath string, opts FsckOptions) (string, error) {
  fsckOpts = opts

  repo, err := openRepo(repoPath)
  if err != nil {
    return "", err
  }

  var cancellable *glib.GCancellable
  var hashIter *glib.GHashTableIter
  var key, value C.gpointer
  var nPartial uint
  var objects, commits *glib.GHashTable
  var tombstones *glib.GPtrArray
  var cerr *C.GError
  defer C.free(unsafe.Pointer(cerr))
  var buffer bytes.Buffer

  if !glib.GoBool(glib.GBoolean(C.ostree_repo_list_objects(repo.native(), C.OSTREE_REPO_LIST_OBJECTS_ALL, (**C.GHashTable)(objects.Ptr()), ((*C.GCancellable)(cancellable.Ptr()), &cerr))) {
    return "", glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
  }

  commits = (*glib.GHashTable)(unsafe.Pointer(C.g_hash_table_new_full((*[0]byte)(C.ostree_hash_object_name), (*[0]byte)(C.g_variant_equal), (C.GDestroyNotify)(C.g_variant_unref), nil)))

  C.g_hash_table_iter_init((*C.GHashTableIter)(hashIter.Ptr()), (*C.GHashTable)(objects.Ptr()))

  if fsckOpts.AddTombstones {
    tombstones = (*glib.GPtrArray)(unsafe.Pointer(C.g_ptr_array_new_with_free_func((*[0]byte)(C.g_free))))
  }

  for glib.GoBool(glib.GBoolean(C.g_hash_table_iter_next((*C.GHashTableIter)(hashIter.Ptr()), &key, &value))) {
    var serializedKey *C.GVariant
    defer C.free(unsafe.Pointer(serializedKey))
    var checksum *C.char
    defer C.free(unsafe.Pointer(checksum))
    var objType C.OstreeObjectType
    var commitState C.OstreeRepoCommitState
    var commit *C.GVariant
    defer C.free(unsafe.Pointer(commit))

    C.ostree_object_name_deserialize(serializedKey, &checksum, &objType)

    if objType == C.OSTREE_OBJECT_TYPE_COMMIT {
      if !glib.GoBool(glib.GBoolean(C.ostree_repo_load_commit(repo.native(), checksum, &commit, &commitState, &cerr))) {
        return "", glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
      }

      if fsckOpts.AddTombstones {
        var parent *C.char = (*C.char)(C.ostree_commit_get_parent(commit))
        defer C.g_free(parent)
        if parent != nil {
          var parentCommit *C.GVariant
          defer C.free(unsafe.Pointer(parentCommit))

          if !glib.GoBool(glib.GBoolean(C.ostree_repo_load_variant(repo.native(), C.OSTREE_OBJECT_TYPE_COMMIT, parent, &parentCommit, &cerr))) {
            if glib.GoBool(glib.GBoolean(C.g_error_matches(cerr, C.g_io_error_quark(), C.G_IO_ERROR_NOT_FOUND))) {
              C.g_ptr_array_add((*C.GPtrArray)(tombstones.Ptr()), C.strdup(checksum))
            } else {
              return "", glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
            }
          }
        }
      }

      if (commitState & C.OSTREE_REPO_COMMIT_STATE_PARTIAL) != 0 {
        nPartial++
      } else {
        C.g_hash_table_insert((*C.GHashTable)(commits.Ptr()), C.g_variant_ref(serializedKey), serializedKey)
      }
    }
  }

  if !fsckOpts.Quiet {
    buffer.WriteString(fmt.Sprintf("Verifying content integrity of %d commit objects...\n", (uint)(C.g_hash_table_size((*C.GHashTable)(commits.Ptr())))))
  }

  foundCorruption, err := fsckReachableObjectsFromCommits(repo, commits, cancellable)
  if err != nil {
    return "", err
  }

  if fsckOpts.AddTombstones {
    var i int
    if tombstones.Length() != 0 {
      if err = enableTombstoneCommits(repo); err != nil {
        return"", err
      }
    }
    
    for i = 0; i < tombstones.Length(); i++ {
      var checksum *C.char =  C.CString(tombstones.PData(i))
      defer C.g_free(checksum)
      buffer.WriteString(fmt.Sprintf("Adding tombostone for commit %s\n", checksum))
      if !glib.GoBool(glib.GBoolean(C.ostree_repo_delete_object(repo.native(), C.OSTREE_OBJECT_TYPE_COMMIT, checksum, (*C.GCancellable)(cancellable.Ptr()), &cerr))) {
        return "", glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
      }
    }
  } else if nPartial > 0 {
    buffer.WriteString(fmt.Sprintf("%d partial commits not verified\n", nPartial))
  }

  if foundCorruption {
    return "", errors.New("Repository corruption encountered")
  }

  return buffer.String(), nil
}

func fsckReachableObjectsFromCommits(repo *Repo, commits *glib.GHashTable, cancellable *glib.GCancellable) (bool, error) {
  var hashIter *glib.GHashTableIter
  var key, value C.gpointer
  var reachableObjects *glib.GHashTable
  var cerr *C.GError

  reachableObjects = glib.ToGHashTable(unsafe.Pointer(C.ostree_repo_traverse_new_reachable()))

  C.g_hash_table_iter_init((*C.GHashTableIter)(hashIter.Ptr()), (*C.GHashTable)(commits.Ptr()))
  for glib.GoBool(glib.GBoolean(C.g_hash_table_iter_next((*C.GHashTableIter)(hashIter.Ptr()), &key, &value))) {
    var serializedKey *glib.GVariant = (*glib.GVariant)(unsafe.Pointer(key))
    var checksum *C.char
    defer C.g_free(checksum)
    var objType C.OstreeObjectType

    C.ostree_object_name_deserialize((*C.GVariant)(serializedKey.Ptr()), &checksum, &objType)

    if objType != C.OSTREE_OBJECT_TYPE_COMMIT {
      return false, errors.New("Expected object type to be a commit")
    }

    if !glib.GoBool(glib.GBoolean(C.ostree_repo_traverse_commit_union(repo.native(), checksum, 0, (*C.GHashTable)(reachableObjects.Ptr()), (*C.GCancellable)(cancellable.Ptr()), &cerr))) {
      return false, glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
    }
  }

  C.g_hash_table_iter_init((*C.GHashTableIter)(hashIter.Ptr()), (*C.GHashTable)(reachableObjects.Ptr()))
  for glib.GoBool(glib.GBoolean(C.g_hash_table_iter_next((*C.GHashTableIter)(hashIter.Ptr()), &key, &value))) {
     var serializedKey *glib.GVariant = (*glib.GVariant)(unsafe.Pointer(key))
     var checksum *C.char
     defer C.g_free(checksum)
     var objType C.OstreeObjectType

     C.ostree_object_name_deserialize((*C.GVariant)(serializedKey.Ptr()), &checksum, &objType)

     foundCorruption := false
     if foundCorruption, err := loadAndFsckOneObject(repo, checksum, objType, cancellable); err != nil {
       return foundCorruption, err
     }

     return foundCorruption, nil
  }
  return false, nil
}

func loadAndFsckOneObject(repo *Repo, checksum *C.char, objType C.OstreeObjectType, cancellable *glib.GCancellable) (bool, error) {
  var missing bool
  var metadata, xattrs *glib.GVariant
  var input *glib.GInputStream
  var fileInfo *glib.GFileInfo
  var cerr *C.GError
  defer C.g_free(unsafe.Pointer(cerr))

  if glib.GoBool(glib.GBoolean(C._ostree_object_type_is_meta(objType))) {
    if !glib.GoBool(glib.GBoolean(C.ostree_repo_load_variant(repo.native(), objType, checksum, (**C.GVariant)(metadata.Ptr()), &cerr))) {
      if !glib.GoBool(glib.GBoolean(C.g_error_matches(cerr, C.g_io_error_quark(), C.G_IO_ERROR_NOT_FOUND))) {
        return false, errors.New(fmt.Sprintf("%s loading metadata object %s", glib.ToGError(unsafe.Pointer(cerr)), C.GoString(checksum)))
      }
      cerr = nil
    } else {
      if objType == C.OSTREE_OBJECT_TYPE_COMMIT {
        if !glib.GoBool(glib.GBoolean(C.ostree_validate_structureof_commit((*C.GVariant)(metadata.Ptr()), &cerr))) {
          return false, errors.New(fmt.Sprintf("%s while validating commit metadata %s", glib.ToGError(unsafe.Pointer(cerr)), C.GoString(checksum)))
  
        }
      } else if objType == C.OSTREE_OBJECT_TYPE_DIR_TREE {
        if !glib.GoBool(glib.GBoolean(C.ostree_validate_structureof_dirtree((*C.GVariant)(metadata.Ptr()), &cerr))) {
          return false, errors.New(fmt.Sprintf("%s while validating directory tree %s", glib.ToGError(unsafe.Pointer(cerr)), C.GoString(checksum)))
        }
      } else if objType == C.OSTREE_OBJECT_TYPE_DIR_META {
        if !glib.GoBool(glib.GBoolean(C.ostree_validate_structureof_dirmeta((*C.GVariant)(metadata.Ptr()), &cerr))) {
          return false, errors.New(fmt.Sprintf("%s while validating directory metadata %s", glib.ToGError(unsafe.Pointer(cerr)), C.GoString(checksum)))
        }
      }

      input = glib.ToGInputStream(unsafe.Pointer(C.g_memory_input_stream_new_from_data(C.g_variant_get_data((*C.GVariant)(metadata.Ptr())), (C.gssize)(C.g_variant_get_size((*C.GVariant)(metadata.Ptr()))), nil)))
    }
  } else {
    var mode uint32
    if objType != C.OSTREE_OBJECT_TYPE_FILE {
      return false, errors.New("Expecting object tpye file")
    }

    if !glib.GoBool(glib.GBoolean(C.ostree_repo_load_file(repo.native(), checksum, (**C.GInputStream)(input.Ptr()), (**C.GFileInfo)(fileInfo.Ptr()), (**C.GVariant)(xattrs.Ptr()), (*C.GCancellable)(cancellable.Ptr()), &cerr))) {
      if !glib.GoBool(glib.GBoolean(C.g_error_matches(cerr, C.g_io_error_quark(), C.G_IO_ERROR_NOT_FOUND))) {
        return false, glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
      }
    } else {
      mode = (uint32)(C.g_file_info_get_attribute_uint32((*C.GFileInfo)(fileInfo.Ptr()), C.CString("unix::mode")))
      if !glib.GoBool(glib.GBoolean(C.ostree_validate_structureof_file_mode((C.guint32)(mode), &cerr))) {
        return false, errors.New(fmt.Sprintf("%s while validating file %s", glib.ToGError(unsafe.Pointer(cerr)), C.GoString(checksum)))
      }
    }
  }

  if !missing {
    var computedCSum *C.guchar
    defer C.g_free(computedCSum)
    var tmpChecksum *C.char
    defer C.g_free(tmpChecksum)

    if !glib.GoBool(glib.GBoolean(C.ostree_checksum_file_from_input((*C.GFileInfo)(fileInfo.Ptr()), (*C.GVariant)(xattrs.Ptr()), (*C.GInputStream)(input.Ptr()), objType, &computedCSum, (*C.GCancellable)(cancellable.Ptr()), &cerr))) {
      return false, glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
    }

    tmpChecksum = C.ostree_checksum_from_bytes(computedCSum)
    if strings.Compare(C.GoString(checksum), C.GoString(tmpChecksum)) != 0 {
      msg := fmt.Sprintf("corrupted object %s.%s; actual checksum: %s", checksum, C.ostree_object_type_to_string(objType), tmpChecksum)

      if fsckOpts.Delete {
        fmt.Println(msg)
        C.ostree_repo_delete_object(repo.native(), objType, checksum, (*C.GCancellable)(cancellable.Ptr()), nil)
        return true, nil
      } else {
        return false, errors.New(msg)
      }
    }
  }

  return false, nil
}
