package otbuiltin

import (
       "unsafe"

       glib "github.com/14rcole/ostree-go/pkg/glibobject"
)

// #cgo pkg-config: ostree-1
// #include <stdlib.h>
// #include <glib.h>
// #include <ostree.h>
// #include "builtin.go.h"
import "C"

type Repo struct {
  *glib.GObject
}

func (r *Repo) native() *C.OstreeRepo {
  return (*C.OstreeRepo)(r.Ptr())
}

func repoFromNative(p *C.OstreeRepo) *Repo {
  if p == nil {
    return nil
  }
  o := (*glib.GObject)(unsafe.Pointer(p))
  r := &Repo{o}
  return r
}

func openRepo(path string) (*Repo, err) {
  var cerr *C.GError = nil
	cpath := C.CString(path)
	pathc := C.g_file_new_for_path(cpath);
	defer C.g_object_unref(C.gpointer(pathc))
	crepo := C.ostree_repo_new(pathc)
	repo := repoFromNative(crepo);
	r := glib.GoBool(glib.GBoolean(C.ostree_repo_open(repo.native(), nil, &cerr)))
	if !r {
		return nil, glibobject.ConvertGError(cerr)
	}
	return repo, nil
}
