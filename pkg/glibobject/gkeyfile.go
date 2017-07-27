/*
 * Copyright (c) 2013 Conformal Systems <info@conformal.com>
 * Copyright (c) 2017 Collabora Ltd <info@collabora.com>
 *
 * This file originated from: http://opensource.conformal.com/
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package glibobject

// #cgo pkg-config: glib-2.0
// #include <glib.h>
import "C"
import (
	"runtime"
	"unsafe"
)

/*
 * GKeyFile
 */

type GKeyFile struct {
	keyfile *C.struct__GKeyFile
}

func (kf *GKeyFile) Native() uintptr {
	if kf == nil || kf.keyfile == nil {
		return uintptr(unsafe.Pointer(nil))
	}
	return uintptr(unsafe.Pointer(kf.keyfile))
}

func (kf *GKeyFile) Ref() {
	C.g_key_file_ref(kf.keyfile)
}

func (kf *GKeyFile) Unref() {
	C.g_key_file_unref(kf.keyfile)
}

func WrapGKeyFile(keyfile uintptr) *GKeyFile {
	ckf := (*C.struct__GKeyFile)(unsafe.Pointer(keyfile))
	if ckf == nil {
		return nil
	}
	kf := &GKeyFile{ckf}
	runtime.SetFinalizer(kf, (*GKeyFile).Unref)

	return kf
}
