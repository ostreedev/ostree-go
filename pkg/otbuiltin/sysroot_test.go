package otbuiltin

import (
	"io/ioutil"
	"os"
	_ "os"
	"path"
	"testing"
)

func TestSysrootDeployment(t *testing.T) {
	osname := "testos"
	refspec := "testbranch"

	baseDir, err := ioutil.TempDir("", "otbuiltin-test-")
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	defer os.RemoveAll(baseDir)

	sysrootDir := path.Join(baseDir, "sysroot")
	os.Mkdir(sysrootDir, 0777)

	commitDir := path.Join(baseDir, "commit1")
	os.Mkdir(commitDir, 0777)

	sysroot := NewSysroot(sysrootDir)
	err = sysroot.InitializeFS()
	if err != nil {
		t.Errorf("%s", err)
	}

	// Try again this time it should fail as it's already initialized
	err = sysroot.InitializeFS()
	if err == nil {
		t.Errorf("Initialized succeeded a second time!", err)
	}

	err = sysroot.InitOsname(osname, nil)
	if err != nil {
		t.Errorf("%s", err)
	}

	repo, err := sysroot.Repo(nil)
	if err != nil {
		t.Fatalf("%s", err)
	}

	/* Dummy commit */
	err = os.Mkdir(path.Join(commitDir, "boot"), 0777)
	if err != nil {
		t.Fatalf("%s", err)
	}

	err = ioutil.WriteFile(path.Join(commitDir, "boot", "vmlinuz-00e3261a6e0d79c329445acd540fb2b07187a0dcf6017065c8814010283ac67f"), []byte("A kernel it is"), 0777)
	if err != nil {
		t.Fatalf("%s", err)
	}

	usretcDir := path.Join(commitDir, "usr", "etc")
	err = os.MkdirAll(usretcDir, 0777)
	if err != nil {
		t.Fatalf("%s", err)
	}

	err = ioutil.WriteFile(path.Join(usretcDir, "os-release"), []byte("PRETTY_NAME=testos\n"), 0777)
	if err != nil {
		t.Fatalf("%s", err)
	}

	opts := NewCommitOptions()
	_, err = repo.PrepareTransaction()
	if err != nil {
		t.Fatalf("%s", err)
	}

	_, err = repo.Commit(commitDir, refspec, opts)
	if err != nil {
		t.Fatalf("%s", err)
	}

	_, err = repo.CommitTransaction()
	if err != nil {
		t.Fatalf("%s", err)
	}

	err = repo.RemoteAdd("origin", "https://example.com", nil, nil)
	if err != nil {
		t.Fatalf("%s", err)
	}

	/* Required by ostree to make sure a bunc of information was pulled in  */
	sysroot.Load(nil)

	revision, err := repo.ResolveRev(refspec, false)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if revision == "" {
		t.Fatal("Revision doesn't exist")
	}

	origin := sysroot.OriginNewFromRefspec(refspec)
	deployment, err := sysroot.DeployTree(osname, revision, origin, nil, nil, nil)
	if err != nil {
		t.Fatalf("%s", err)
	}

	err = sysroot.SimpleWriteDeployment(osname, deployment, nil, 0, nil)
	if err != nil {
		t.Fatalf("%s", err)
	}
}
