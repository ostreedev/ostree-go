package otbuiltin

import (
  "testing"
  "os"
  //"strconv"
  "fmt"

  //"github.com/14rcole/gopopulate"
)

func TestPruneNoPrunePass(t *testing.T) {
  // Create a temporary repository
  /*baseDir := "/tmp/ostree-go-test"
  err := os.Mkdir(baseDir, 0777)
  if err != nil {
    t.Errorf("%s", err)
    return
  }
  defer os.RemoveAll(baseDir)*/
  //repoDir := baseDir + "/repo"
  repoDir := "/tmp/test-init-repo"
  fmt.Println(repoDir)
  err := os.Mkdir(repoDir, 0777)
  if err != nil {
    t.Errorf("%s", err)
    return
  }
  defer os.RemoveAll(repoDir)

  // Initialize the repo
  inited, err := Init(repoDir, nil)
  if err != nil {
    t.Errorf("%s", err)
    return
  } else if !inited {
    t.Errorf("Cannot test commit: failed to initialize repo")
    return
  }

  // Let's make a few commits
  /*commitDir := baseDir + "/commits"
  err = os.Mkdir(commitDir, 0777)
  if err != nil {
    t.Errorf("%s", err)
    return
  }
  for i := 0; i < 5; i++ {
    newDir := commitDir + strconv.Itoa(i)
    err = os.Mkdir(newDir, 0777)
    if err != nil {
      t.Errorf("%s", err)
      return
    }
    err = gopopulate.PopulateDir(newDir, "rd", 3, 4)
    if err != nil {
      t.Errorf("%s", err)
      return
    }
    commitOpts := NewCommitOptions()
    branch := "ot-test-commit-" + strconv.Itoa(i)
    _, err := Commit(repoDir, newDir, branch, commitOpts)
    if err != nil {
      t.Errorf("%s", err)
      return
    }
    fmt.Println("Commits completed: ", i)
  }*/

  //Test commit
  opts := NewCommitOptions()
  branch := "test-branch"
  ret, err := Commit(repoDir, "/home/rycole/Development/C-C++/ostree/", branch, opts)
  if err != nil {
    t.Errorf("%s", err)
  } else {
    fmt.Println(ret)
  }

  // Now let's do some pruning!
  pruneOpts := NewPruneOptions()
  pruneOpts.NoPrune = true
  ret, err = Prune(repoDir, pruneOpts)
  if err != nil {
    t.Errorf("%s", err)
  } else {
    fmt.Println(ret)
  }
}

func TestPruneNoPruneFail(t *testing.T) {
}

func TestPrunePass(t *testing.T) {
}

func TestPruneFail(t *testing.T) {
}
