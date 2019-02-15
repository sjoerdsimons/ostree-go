package otbuiltin

import (
	"bytes"
	"fmt"
	glib "github.com/sjoerdsimons/ostree-go/pkg/glibobject"
	"reflect"
	"strings"
	"unicode"
	"unsafe"
)

// #cgo pkg-config: ostree-1
// #include <stdlib.h>
// #include <glib.h>
// #include <ostree.h>
// #include "builtin.go.h"
import "C"

type RemoteOptions struct {
	ContentUrl          string "contenturl"
	Proxy               string
	NoGpgVerify         bool "gpg-verify, invert"
	NoGpgVerifySummary  bool "gpg-verify-summary, invert"
	TlsPermissive       bool
	TlsClientCertPath   string
	TlsClientKeyPath    string
	TlsCaPath           string
	UnconfiguredState   string
	MinFreeSpacePercent string
	CollectionId        string
}

func toDashString(in string) string {
	var out bytes.Buffer
	for i, c := range in {
		if !unicode.IsUpper(c) {
			out.WriteRune(c)
			continue
		}

		if i > 0 {
			out.WriteString("-")
		}
		out.WriteRune(unicode.ToLower(c))
	}

	return out.String()
}

func optionsToVariant(options RemoteOptions) *C.GVariant {
	casv := C.CString("a{sv}")
	defer C.free(unsafe.Pointer(casv))
	csv := C.CString("{sv}")
	defer C.free(unsafe.Pointer(csv))

	builder := C.g_variant_builder_new(C._g_variant_type(casv))

	v := reflect.ValueOf(options)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		vf := v.Field(i)
		tf := t.Field(i)
		invert := false

		name := toDashString(tf.Name)
		if tf.Tag != "" {
			opts := strings.Split(string(tf.Tag), ",")
			if opts[0] != "" {
				name = opts[0]
			}
			for _, o := range opts[1:] {
				switch strings.TrimSpace(o) {
				case "invert":
					invert = true
				default:
					panic(fmt.Sprintf("Unhandled flag: %s", o))
				}
			}
		}

		var variant *C.GVariant
		switch vf.Kind() {
		case reflect.Bool:
			/* Should probalby use e.g. Maybe type so it can judge unitialized */
			b := vf.Bool()
			var cb C.gboolean

			if !b {
				// Still the default, so don't bother setting it
				continue
			}
			if invert {
				cb = C.gboolean(0)
			} else {
				cb = C.gboolean(1)
			}
			variant = C.g_variant_new_boolean(cb)
		case reflect.String:
			if vf.String() == "" {
				continue
			}
			variant = C.g_variant_new_take_string((*C.gchar)(C.CString(vf.String())))
		default:
			panic(fmt.Sprintf("Can't handle type of field: %s", tf.Name))
		}

		cname := C.CString(name)
		defer C.free(unsafe.Pointer(cname))
		C._g_variant_builder_add_twoargs(builder, csv, cname, variant)
	}

	coptions := C.g_variant_builder_end(builder)
	return coptions
}

func (repo *Repo) RemoteAdd(name, url string, options RemoteOptions,
	cancellable *glib.GCancellable) error {

	var cerr *C.GError = nil

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))

	coptions := optionsToVariant(options)
	C.g_variant_ref_sink(coptions)

	r := C.ostree_repo_remote_add(repo.native(), cname, curl, coptions, cCancellable(cancellable), &cerr)

	C.g_variant_unref(coptions)

	if !gobool(r) {
		return generateError(cerr)
	}

	return nil
}
