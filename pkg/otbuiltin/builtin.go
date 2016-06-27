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
  o := glib.GObjectNew(unsafe.Pointer(p))
  r := &Repo{o}
  return r
}
