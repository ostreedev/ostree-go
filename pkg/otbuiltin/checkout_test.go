package otbuiltin

import (
       "testing"
       _ "os"
       _ "fmt"
       _ "time"
)

func TestCheckoutSuccessProcessOneBranch(t *testing.T) {
  // Make a directory in which the repo should exist
  /*testDir := "/tmp/test-init-repo"
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
  commitOpts := NewCommitOptions()
  branch := "test-branch"
  ret, err := Commit(testDir, ".", branch, commitOpts)
  if err != nil {
    t.Errorf("%s", err)
  } else {
    fmt.Println(ret)
  }

  fmt.Println("Sleeping so you can check the repo manually...")
  d, _ := time.ParseDuration("30s")
  time.Sleep(d)

  checkoutOpts := NewCheckoutOptions()
  checkoutDir := "/tmp/checkout"
  err := os.Mkdir(checkoutDir, 0777)
  if err != nil {
    t.Errorf("%s", err)
    return
  }
  defer os.RemoveAll(checkoutDir)*/

  //err = Checkout(testDir, checkoutDir, branch, checkoutOpts)
  branch := "ot-branch"
  defer os.RemoveAll(branch)
  err := Checkout("/tmp/repo", "/tmp/ot-branch", "ot-branch", checkoutOpts)
  if err != nil {
    t.Errorf("%s", err)
    return
  }
}

func TestCheckoutSuccessProcessOneCommit(t *testing.T) {
  
}

func TestCheckoutFailProcessOne(t *testing.T) {

}

func TestCheckoutSuccessProcessMany(t *testing.T) {

}

func TestCheckoutFailProcessMany(t *testing.T) {

}
