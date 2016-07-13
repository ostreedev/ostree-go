package otbuiltin

import (
       "testing"
        "os"
        "fmt"
        //"time"
)

func TestCheckoutSuccessProcessOneBranch(t *testing.T) {
  // Make a directory in which the repo should exist
  repoDir := "/tmp/test-init-repo"
  err := os.Mkdir(repoDir, 0777)
  if err != nil {
    t.Errorf("%s", err)
    return
  }
  defer os.RemoveAll(repoDir)

  // Initialize the repo
  inited, err := Init("/tmp/test-init-repo", nil)
  if !inited || err != nil {
    fmt.Println("Cannot test commit: failed to initialize repo")
    return
  }

  //Commit to the repo
  commitOpts := NewCommitOptions()
  branch := "test-branch"
  ret, err := Commit(repoDir, "/home/rycole/Development/C-C++/ostree", branch, commitOpts)
  if err != nil {
    t.Errorf("%s", err)
  } else {
    fmt.Println(ret)
  }

  checkoutOpts := NewCheckoutOptions()
  checkoutDir := "/tmp/checkout"
  if err != nil {
    t.Errorf("%s", err)
    return
  }
  // The directory will be created when the Checkout is made
  defer os.RemoveAll(checkoutDir)

  /*fmt.Println("This is your opportunity to do a quick checkout yourself")
  d, _ := time.ParseDuration("30s")
  time.Sleep(d)
*/
  err = Checkout(repoDir, checkoutDir, branch, checkoutOpts)
  defer os.RemoveAll(branch)
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
