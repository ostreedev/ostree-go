package main

import (
  "fmt"
  "os"
  o "github.com/14rcole/ostree-go/pkg/otbuiltin"

  "github.com/14rcole/gopopulate"
)

func main() {
	// Make a base directory in which all of our test data resides
	baseDir := "/tmp/otbuiltin-test/"
	err := os.Mkdir(baseDir, 0777)
	if err != nil {
		fmt.Errorf("%s", err)
		return
	}
	defer os.RemoveAll(baseDir)
	// Make a directory in which the repo should exist
	repoDir := baseDir + "repo"
	err = os.Mkdir(repoDir, 0777)
	if err != nil {
		fmt.Errorf("%s", err)
		return
	}

	// Initialize the repo
  initOpts := o.NewInitOptions()
  initOpts.Mode = "bare-user"
	inited, err := o.Init(repoDir, initOpts)
	if !inited || err != nil {
		fmt.Println("Cannot test commit: failed to initialize repo")
		return
  }

	//Make a new directory full of random data to commit
	commitDir := baseDir + "commit1"
	err = os.Mkdir(commitDir, 0777)
	if err != nil {
		fmt.Errorf("%s", err)
		return
	}
	err = gopopulate.PopulateDir(commitDir, "rd", 4, 4)
	if err != nil {
		fmt.Errorf("%s", err)
		return
	}

	//Test commit
	opts := o.NewCommitOptions()
	branch := "test-branch"
	ret, err := o.Commit(repoDir, commitDir, branch, opts)
	if err != nil {
		fmt.Errorf("%s", err)
	} else {
		fmt.Println(ret)
	}
}
