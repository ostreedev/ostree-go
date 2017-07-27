/*
 * Copyright (c) 2013 Conformal Systems <info@conformal.com>
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

// #cgo pkg-config: glib-2.0 gobject-2.0
// #include <glib.h>
// #include <glib-object.h>
// #include <gio/gio.h>
// #include "glibobject.go.h"
// #include <stdlib.h>
import "C"
import (
	"unsafe"
)

/*
 * GObject
 */

// ToGObject type converts an unsafe.Pointer as a native C GObject.
// This function is exported for visibility in other gotk3 packages and
// is not meant to be used by applications.
func ToGObject(p unsafe.Pointer) *C.GObject {
	return C.toGObject(p)
}

// IObject is an interface type implemented by Object and all types which embed
// an Object.  It is meant to be used as a type for function arguments which
// require GObjects or any subclasses thereof.
type IObject interface {
	toGObject() *C.GObject
	ToObject() *Object
}

// GObject is a representation of GLib's GObject.
type Object struct {
	GObject *C.GObject
}

func (v *Object) Native() uintptr {
	if v == nil || v.GObject == nil {
		return uintptr(unsafe.Pointer(nil))
	}
	return uintptr(unsafe.Pointer(v.GObject))
}

func (v *Object) Ref() {
	C.g_object_ref(C.gpointer(v.GObject))
}

func (v *Object) Unref() {
	C.g_object_unref(C.gpointer(v.GObject))
}

func (v *Object) RefSink() {
	C.g_object_ref_sink(C.gpointer(v.GObject))
}

func (v *Object) IsFloating() bool {
	c := C.g_object_is_floating(C.gpointer(v.GObject))
	return GoBool(GBoolean(c))
}

func (v *Object) ForceFloating() {
	C.g_object_force_floating(v.GObject)
}
