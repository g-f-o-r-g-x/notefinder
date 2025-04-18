package main

/*
#cgo pkg-config: poppler-glib
#include <poppler.h>
*/
import "C"

import (
	"strings"
	"unsafe"
)

func pdfMatchesPattern(uri string, pattern string) bool {
	cUri := C.CString(uri)
	defer C.free(unsafe.Pointer(cUri))

	doc := C.poppler_document_new_from_file(cUri, nil, nil)
	defer C.g_object_unref(C.gpointer(doc))
	if doc == nil {
		return false
	}

	for pageNum := range int(C.poppler_document_get_n_pages(doc)) {
		page := C.poppler_document_get_page(doc, C.int(pageNum))
		defer C.g_object_unref(C.gpointer(page))
		pageText := C.GoString(C.poppler_page_get_text(page))
		if strings.Contains(
			strings.ToLower(pageText),
			strings.ToLower(pattern)) {
			return true
		}

	}

	return false
}
