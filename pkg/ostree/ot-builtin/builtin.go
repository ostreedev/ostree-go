package ot-builtin

import (
       "unsafe"
)

// #cgo pkg-config: ostree-1
// #include <stdlib.h>
// #include <glib.h>
// #include <ostree.h>
// #include "builtin.go.h"
import "C"

type Repo struct {
  *GObject
}

func (r *Repo) native() *C.OstreeRepo {
  return (*C.OstreeRepo)(r.ptr)
}

func repoFromNative(p *OstreeRepo) {
  if p == nil {
    return nil
  }
  o := GObjectNew(unsafe.Pointer(p))
  r := &Repo{o}
  return r
}
