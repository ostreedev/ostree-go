package otbuiltin

import (
       "testing"
       "os"
)

func TestInit(t *testing.T) {
  // Create an empty directory is we know it's not a repo
  testDir := "/tmp/test-init-repo"
  err := os.Mkdir(testDir, 0777)
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

  os.Remove(testDir)
}
