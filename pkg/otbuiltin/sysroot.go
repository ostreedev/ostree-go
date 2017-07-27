package otbuiltin

import (
	"errors"
	glib "github.com/ostreedev/ostree-go/pkg/glibobject"
	"os"
	"path"
	"runtime"
	"unsafe"
)

// #cgo pkg-config: ostree-1
// #include <stdlib.h>
// #include <glib.h>
// #include <ostree.h>
// #include "builtin.go.h"
import "C"

type Sysroot struct {
	*glib.Object
	path string
}

func NewSysroot(path string) *Sysroot {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	cgfile := C.g_file_new_for_path(cpath)
	defer C.g_object_unref(C.gpointer(cgfile))

	s := C.ostree_sysroot_new(cgfile)
	g := glib.ToGObject(unsafe.Pointer(s))
	obj := &glib.Object{g}
	sysroot := &Sysroot{obj, path}

	runtime.SetFinalizer(sysroot, (*Sysroot).Unref)

	return sysroot
}

func (s *Sysroot) native() *C.OstreeSysroot {
	if s == nil || s.GObject == nil {
		return nil
	}
	return (*C.OstreeSysroot)(unsafe.Pointer(s.GObject))
}

func (s Sysroot) Path() string {
	return s.path
}

func (s Sysroot) InitializeFS() error {
	toplevels := []struct {
		name string
		perm os.FileMode
	}{
		{"boot", 0777},
		{"dev", 0777},
		{"home", 0777},
		{"proc", 0777},
		{"run", 0777},
		{"sys", 0777},
		{"tmp", 01777},
		{"root", 0700},
	}

	if _, err := os.Stat(path.Join(s.path, "ostree")); err == nil {
		return errors.New("Filesystem already initialized for ostree")
	}

	for _, t := range toplevels {
		p := path.Join(s.path, t.name)
		err := os.Mkdir(p, t.perm)
		if err != nil && !os.IsExist(err) {
			return err
		}
	}

	return s.EnsureInitialized(nil)
}

func (s *Sysroot) EnsureInitialized(cancellable *glib.GCancellable) error {
	var cerr *C.GError = nil

	r := C.ostree_sysroot_ensure_initialized(s.native(), cCancellable(cancellable), &cerr)

	if !gobool(r) {
		return generateError(cerr)
	}
	return nil
}

func (s *Sysroot) InitOsname(name string, cancellable *glib.GCancellable) error {
	var cerr *C.GError = nil
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	r := C.ostree_sysroot_init_osname(s.native(), cname, cCancellable(cancellable), &cerr)

	if !gobool(r) {
		return generateError(cerr)
	}
	return nil
}

func (s *Sysroot) Load(cancellable *glib.GCancellable) error {
	var cerr *C.GError = nil
	r := C.ostree_sysroot_load(s.native(), cCancellable(cancellable), &cerr)

	if !gobool(r) {
		return generateError(cerr)
	}
	return nil
}

func (s *Sysroot) Repo(cancellable *glib.GCancellable) (*Repo, error) {
	var repo *C.OstreeRepo
	var cerr *C.GError = nil

	r := C.ostree_sysroot_get_repo(s.native(), &repo, cCancellable(cancellable), &cerr)

	if !gobool(r) {
		return nil, generateError(cerr)
	}
	return repoFromNative(repo), nil
}

func (s *Sysroot) OriginNewFromRefspec(refspec string) *glib.GKeyFile {
	crefspec := C.CString(refspec)
	defer C.free(unsafe.Pointer(crefspec))

	kf := C.ostree_sysroot_origin_new_from_refspec(s.native(), crefspec)
	return glib.WrapGKeyFile(uintptr(unsafe.Pointer(kf)))
}

func (s *Sysroot) PrepareCleanup(cancellable *glib.GCancellable) error {
	var cerr *C.GError = nil
	r := C.ostree_sysroot_prepare_cleanup(s.native(), cCancellable(cancellable), &cerr)

	if !gobool(r) {
		return generateError(cerr)
	}
	return nil
}

func (s *Sysroot) DeployTree(osname, revision string, origin *glib.GKeyFile, providedMergeDeployment *Deployment, overrideKernelArgv []string, cancellable *glib.GCancellable) (*Deployment, error) {
	var cerr *C.GError = nil
	var cdeployment *C.OstreeDeployment = nil

	cosname := C.CString(osname)
	defer C.free(unsafe.Pointer(cosname))

	crevision := C.CString(revision)
	defer C.free(unsafe.Pointer(crevision))

	coverrideKernelArgv := make([]*C.char, len(overrideKernelArgv)+1)
	for i, s := range overrideKernelArgv {
		coverrideKernelArgv[i] = C.CString(s)
	}
	coverrideKernelArgv[len(overrideKernelArgv)] = nil

	r := C.ostree_sysroot_deploy_tree(s.native(), cosname, crevision,
		(*C.struct__GKeyFile)(unsafe.Pointer(origin.Native())),
		providedMergeDeployment.native(),
		(**C.char)(unsafe.Pointer(&coverrideKernelArgv[0])),
		&cdeployment,
		cCancellable(cancellable), &cerr)

	for i, s := range coverrideKernelArgv {
		coverrideKernelArgv[i] = nil
		C.free(unsafe.Pointer(s))
	}

	if !gobool(r) {
		return nil, generateError(cerr)
	}
	return wrapDeployment(cdeployment), nil
}

type SysrootSimpleWriteDeploymentFlags int

const (
	OSTREE_SYSROOT_SIMPLE_WRITE_DEPLOYMENT_FLAGS_NONE        SysrootSimpleWriteDeploymentFlags = 1 << iota
	OSTREE_SYSROOT_SIMPLE_WRITE_DEPLOYMENT_FLAGS_RETAIN      SysrootSimpleWriteDeploymentFlags = 1 << iota
	OSTREE_SYSROOT_SIMPLE_WRITE_DEPLOYMENT_FLAGS_NOT_DEFAULT SysrootSimpleWriteDeploymentFlags = 1 << iota
	OSTREE_SYSROOT_SIMPLE_WRITE_DEPLOYMENT_FLAGS_NO_CLEAN    SysrootSimpleWriteDeploymentFlags = 1 << iota
)

func (s *Sysroot) SimpleWriteDeployment(osname string, newDeployment,
	mergeDeployment *Deployment, flags SysrootSimpleWriteDeploymentFlags, cancellable *glib.GCancellable) error {
	var cerr *C.GError = nil
	cosname := C.CString(osname)
	defer C.free(unsafe.Pointer(cosname))

	r := C.ostree_sysroot_simple_write_deployment(s.native(),
		cosname,
		newDeployment.native(),
		mergeDeployment.native(),
		(C.OstreeSysrootSimpleWriteDeploymentFlags)(flags),
		cCancellable(cancellable), &cerr)

	if !gobool(r) {
		return generateError(cerr)
	}
	return nil
}
