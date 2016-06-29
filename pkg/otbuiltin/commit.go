package otbuiltin

import (
  glib "github.com/14rcole/ostree-go/pkg/glibobject"
)

func Commit(options map[string]string) {

}

func parseArgs(options map[string]string) {

}


type CommitOptions struct {
  Subject                   string    // One line subject
  Body                      string    // Full description
  Branch                    string    // branch
  Tree                      string    // 'dir=PATH' or 'tar=TARFILE' or 'ref=COMMIT': overlay the given argument as a tree
  AddMetadataString         string    // Add a key/value pair to metadata
  AddDetachedMetadataString string    // Add a key/value pair to detached metadata
  OwnerUID                  string    // Set file ownership to user id
  OwnerGID                  string    // Set file ownership to group id
  NoXattrs                  string    // Do not import extended attributes
  LinkCheckoutSpeedup       bool      // Optimize for commits of trees composed of hardlinks in the repository
  TarAuotocreateParents     bool      // When loading tar archives, automatically create parent directories as needed
  SkipIfUnchanged           bool      // If the contents are unchanged from a previous commit, do nothing
  StatOverride              string    // File containing list of
  TableOutput               bool      // Output more information in a KEY: VALUE format
  GenerateSizes             bool      // Generate size information along with commit metadata
  GpgSign                   []string  // GPG Key ID with which to sign the commit (if you have GPGME - GNU Privacy Guard Made Easy)
  GpgHomedir                string    // GPG home directory to use when looking for keyrings (if you have GPGME - GNU Privacy Guard Made Easy)
}
