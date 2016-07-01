package otbuiltin

import (
       "unsafe"
       "time"

       glib "github.com/14rcole/ostree-go/pkg/glibobject"
)

// #cgo pkg-config: ostree-1
// #include <stdlib.h>
// #include <glib.h>
// #include <ostree.h>
// #include "builtin.go.h"
import "C"

type PruneOptions struct {
  NoPrune           bool      // Only display unreachable objects; don't delete
  RefsOnly          bool      // Only compute reachability via refs
  DeleteCommit      string    // Specify a commit to delete
  KeepYoungerThan   time.Time // All commits older than this date will be pruned
  Depth             int = -1  // Only traverse depths (integer) parents for each commit (default: -1=infinite)
  StaticDeltasOnly  int       // Change the behavior of --keep-younger-than and --delete-commit to prune only the static delta files
}

func Prune(repo string) error{
  return nil
}

func deleteCommit(repo *OstreeRepo, commitToDelete string, cancellable *GCancellable) error {
  return nil
}

func pruneCommitsKeepYoungerThanDate(repo *OstreeRepo, date time.Time, cancellable *GCancellable) error {
  return nil
}
