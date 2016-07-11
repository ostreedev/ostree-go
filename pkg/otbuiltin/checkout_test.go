package otbuiltin

import (
       "testing"
)

func TestCheckoutSuccessProcessOne(t *testing.T) {
  // Make a directory in which the repo should exist
  testDir := "/tmp/test-init-repo"
  err := os.Mkdir(testDir, 0777)
  if err != nil {
    t.Errorf("%s", err)
    return
  }
  defer os.RemoveAll(testDir)

  // Initialize the repo
  inited, err := Init("/tmp/test-init-repo", nil)
  if !inited || err != nil {
    fmt.Println("Cannot test commit: failed to initialize repo")
    return
  }

  //Commit to the repo
  opts := NewCommitOptions()
  branch := "test-branch"
  ret, err = Commit(testDir, ".", branch, opts)
  if err != nil {
    t.Errorf("%s", err)
  } else {
    fmt.Println(ret)
  }

  opts := NewCheckoutOptions()
  checkoutDir := "/tmp/checkout"
  err = os.Mkdir(checkoutDir, 0777)
  if err != nil {
    t.Errorf("%s", err)
    return
  }

  err = Checkout(testDir, checkoutDir, branch, opts)
  if err != nil {
    t.Errorf("%s", err)
    return
  }
}

func TestCheckoutFailProcessOne(t *testing.T) {

}

func TestCheckoutSuccessProcessMany(t *testing.T) {

}

func TestCheckoutFailProcessMany(t *testing.T) {

}
