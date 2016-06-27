package otbuiltin

import (
       "testing"
       "os"

       glib "github.com/14rcole/ostree-go/pkg/glibobject"
)

func TestInit(t *testing.T) {
  // Create an empty directory is we know it's not a repo
  err := os.Mkdir("/tmp/test-init-repo", 0777)
  if (err != nil){
    t.Errorf("%s", err)
    return
  }

  // Try to init the repo
  // In this case, inited should be true and err should be nil
  inited, err := Init("/tmp/test-init-repo", nil)
  if !inited || err != nil {
    t.Errorf("%s", err)
    return
  }

  // Init the repo again
  // Since the repo already exists inited should be true and err should not be nil
  inited, err = Init("tmp/test-init-repo", nil)
  if err == nil {
    t.Errorf("second initialization overwrote repo")
    return
  } else if !inited {
    t.Errorf("%s", err)
    return
  }
}
