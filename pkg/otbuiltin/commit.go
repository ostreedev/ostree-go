package otbuiltin

import (
  "errors"
  "strings"
  "time"
  "unsafe"
  "path"
  "os"
  "fmt"

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
  Subject                   string    // One line subject
  Body                      string    // Full description
  Parent                    string    // Parent of the commit
  Tree                      []string  // 'dir=PATH' or 'tar=TARFILE' or 'ref=COMMIT' or 'layer=TARFILE': overlay the given argument as a tree
  OwnerUID                  int       // Set file ownership to user id
  OwnerGID                  int       // Set file ownership to group id
  TarAutoCreateParents     bool      // When loading tar archives, automatically create parent directories as needed
  Timestamp                 time.Time // Override the timestamp of the commit
  Orphan                    bool      // Commit does not belong to a branch
  Fsync                     bool      // Specify whether fsync should be used or not.  Default to true
}

func NewCommitOptions() commitOptions {
  var co commitOptions
  co.OwnerUID = -1
  co.OwnerGID = -1
  co.Fsync = true
  return co
}

func DockerCommitOptions(layerPath string) commitOptions {
  co := NewCommitOptions()
  co.Tree = append(co.Tree, path.Join("layer=", layerPath))
  co.TarAutoCreateParents = true
  return co
}

// Commit data to an ostree repository
func Commit(repoPath, commitPath, branch string, opts commitOptions) (string, error) {
	options = opts
	
	repo, err := openRepo(repoPath)
	if err != nil {
		return "", err
	}

  if !(strings.Compare(opts.Branch, "") == 0 || opts.Orphan) {
    return "", errors.New("A branch must be specified or set opts.Orphan=true")
  }

  rev, err = resolveParent(repo, opts.Parent, opts.Orphan, opts.Branch)
  if err != nil {
    return "", err
  }

	err = prepareTransaction(repo)
	if err != nil {
		return "", err
	}

	var root *C.GFile
  var modifier *C.OstreeRepoCommitModifier

	err = writeToMtree(repo, modifier, commitPath, options.Tree, root)
	if err != nil {
		return "", err
	}

  var metadata *C.GVariant
  return writeCommit(repo, options.Parent, options.Subject, options.Body, metadata, root)
}

func resolveParent(repo *repo, parent, orphan, branch string) (string, error) {
  var err *C.GError
  defer C.g_free(err)
  if strings.Compare(parent, "") != 0 {
    if strings.Compare(parent, "none") == 0 {
      parent = ""
    } else {
      if !glib.GoBool(glib.GBoolean(C.ostree_validate_checksum_string(C.CString(parent), &err))) {
        return glib.ConvertGError(glib.ToGError(unsafe.Pointer(err)))
      }
    }
  } else if !orphan {
    cparent := C.CString(parent)
    if !glib.GoBool(glib.GBoolean(C.ostree_repo_resolve_rev(repo.native(), C.CString(branch), C.TRUE, &cparent, &err))) {
      return glib.ConvertGError(glib.ToGError(unsafe.Pointer(errr)))
    }
    return C.GoString(cparent), nil
  }
}

func prepareTransaction(repo *Repo) error {
	var cerr *C.GError
	defer C.g_free(unsafe.Pointer(cerr))
	C.ostree_repo_prepare_transaction(repo.native(), nil, nil, &cerr)
	if cerr != nil {
		return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
	}
	return nil
}

func writeToMtree(repo *Repo, modifier *C.OstreeRepoCommitModifier, path string, tree []string, root *C.GFile) error {
	var cerr *C.GError
	defer C.g_free(unsafe.Pointer(cerr))
	var err error
	mtree := mutableTreeFromNative(C.ostree_mutable_tree_new())
	
	if len(path) == 0 && len(tree) == 0 {
		err = writeCwdToMtree(repo, mtree, modifier)
	} else if len(tree) != 0 {
		err = writeTreeToMtree(repo, mtree, modifier, tree)
	} else {
		err = writePathToMtree(repo, mtree, modifier, path)
	}
	if err != nil {
		return err
	}

	if !glib.GoBool(glib.GBoolean(C.ostree_repo_write_mtree(repo.native(), mtree.native(), &root, nil, &cerr))) {
		return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
	}
  return nil
}

func writeCwdToMtree(repo *Repo, mtree *OstreeMutableTree, modifier *C.OstreeRepoCommitModifier) error {
	workingDir, err := os.Getwd()
	if err != nil {
		return err
	}
	
	return writePathToMtree(repo, mtree, modifier, workingDir)
}

func writePathToMtree(repo *Repo, mtree *OstreeMutableTree, modifier *C.OstreeRepoCommitModifier, path string) error {
	var cerr *C.GError
	defer C.g_free(unsafe.Pointer(cerr))
  fmt.Println(repo.isInitialized())

	if !glib.GoBool(glib.GBoolean(C.ostree_repo_write_directory_to_mtree(repo.native(), C.g_file_new_for_path(C.CString(path)), mtree.native(), modifier, nil, &cerr))) {
		return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
	}
	return nil
}

func writeTreeToMtree(repo *Repo, mtree *OstreeMutableTree, modifier *C.OstreeRepoCommitModifier, tree []string) error {
	for i := range(tree) {
		treeType, treeVal, err := split(tree[i])
		if err != nil {
			return err
		}

		if strings.Compare(treeType, "dir") == 0 {
			err = writePathToMtree(repo, mtree, modifier, treeVal)
		} else if strings.Compare(treeType, "tar") == 0 {
			err = writeTarToMtree(repo, mtree, modifier, treeVal)
		} else if strings.Compare(treeType, "layer") == 0 {
			err = writeLayerToMtree(repo, mtree, modifier, treeVal)
		} else if strings.Compare(treeType, "ref") == 0 {
			err = writeRefToMtree(repo, mtree, modifier, treeVal)
		} else {
			return errors.New("Invalid type in tree specification")
		}
		if err != nil {
			return err
		}
 	}
	return nil
}

func writeTarToMtree(repo *Repo, mtree *OstreeMutableTree, modifier *C.OstreeRepoCommitModifier, tarFile string) error {
	var cerr *C.GError
	defer C.g_free(unsafe.Pointer(cerr))

	if !glib.GoBool(glib.GBoolean(C.ostree_repo_write_archive_to_mtree(repo.native(), C.g_file_new_for_path(C.CString(tarFile)), mtree.native(), modifier, (C.gboolean)(glib.GBool(options.TarAutoCreateParents)), nil, &cerr))) {
		return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))	
	}
	
	return nil
}

func writeLayerToMtree(repo *Repo, mtree *OstreeMutableTree, modifier *C.OstreeRepoCommitModifier, layer string) error {
	var cerr *C.GError
	defer C.g_free(unsafe.Pointer(cerr))

	if !glib.GoBool(glib.GBoolean(C.ostree_repo_import_oci_image_layer(repo.native(), nil, -1, C.CString(layer), mtree.native(), modifier, nil, &cerr))) {
		return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))	
	}
	
	return nil
}

func writeRefToMtree(repo *Repo, mtree *OstreeMutableTree, modifier *C.OstreeRepoCommitModifier, ref string) error {
	var cerr *C.GError
	defer C.g_free(unsafe.Pointer(cerr))

	var objectToCommit *glib.GFile

	if !glib.GoBool(glib.GBoolean(C.ostree_repo_read_commit(repo.native(), C.CString(ref), (**C.GFile)(objectToCommit.Ptr()), nil, nil, &cerr))) {
		return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))	
	}

	if !glib.GoBool(glib.GBoolean(C.ostree_repo_write_directory_to_mtree(repo.native(), (*C.GFile)(objectToCommit.Ptr()), mtree.native(), modifier, nil, &cerr))) {
		return glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
	}

	return nil
}


func writeCommit(repo *Repo, parent, subject, body string, metadata *C.GVariant, root *C.GFile) (string, error) {
	var cerr *C.GError
	defer C.g_free(unsafe.Pointer(cerr))
	var checksum *C.char
	defer C.free(unsafe.Pointer(checksum))

  crepo := repo.native()
  cparent := C.CString(parent)
  csubject := C.CString(subject)
  cbody := C.CString(body)
  repoFileRoot := C._ostree_repo_file(root)
	if !glib.GoBool(glib.GBoolean(C.ostree_repo_write_commit(crepo, cparent,csubject, cbody, metadata, repoFileRoot, &checksum, nil, &cerr))) {
		return "", glib.ConvertGError(glib.ToGError(unsafe.Pointer(cerr)))
	}
	return C.GoString(checksum), nil
}



// Split a tree value into it's key/value pair
func split(s string) (string, string, error) {
	pair := strings.SplitN(s, "=", 2)
	if len(pair) < 2 {
		return "", "", errors.New("Tree must contain strings in format KEY=VALUE")
	}

	return pair[0], pair[1], nil
}
