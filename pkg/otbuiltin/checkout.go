package otbuiltin

import (
       "errors"
       "strings"
       "fmt"
       "unsafe"

       glib "github.com/14rcole/ostree-go/pkg/glibobject"
)

// #cgo pkg-config: ostree-1
// #include <stdlib.h>
// #include <glib.h>
// #include <ostree.h>
// #include "builtin.go.h"
import "C"

var options CheckoutOptions

type CheckoutOptions struct {
  UserMode        bool    // Do not change file ownership or initialize extended attributes
  DisableCache    bool    // Do not update or use the internal repository uncompressed object
  Union           bool    // Keep existing directories and unchanged files, overwriting existing filesystem
  AllowNoent      bool    // Do nothing if the specified filepath does not exist
  Subpath         string  // Checkout sub-directory path
  FromFile        string  // Process many checkouts from the given file
}

func Checkout(repo, filepath, commit string, opts CheckoutOptions) error {
  if opts != (CheckoutOptions{}) {
    options = opts
  }
  return nil
}

func processOneCheckout(OstreeRepo *repo, resolved_commit, subpath, destination string, cancellable glib.GCancellable) error {

}

func processManyCheckouts(OstreeRepo *repo, target string, cancellable glib.GCancellable) error {

}
