package main

import (
	"time"
	"unsafe"
)

/*
#include <EXTERN.h>
#include <perl.h>
*/
import "C"

const (
	FlagArchived = 1 << 0
	FlagReadOnly = 1 << 1
	FlagNotify   = 1 << 2
)

type Markup int

const (
	MarkupNone Markup = iota
	Markdown
	MarkupHTML
	MarkupTodoTxt
)

type NoteType int

const (
	NoteTypeRegular NoteType = iota
	NoteTypeBookmark
	NoteTypeVoice
	NoteTypeFile
)

type Note struct {
	context    *Context
	UUID       uint64
	Title      string
	Body       string
	URI        string
	MimeType   string
	CreatedAt  time.Time
	ModifiedAt time.Time
	flags      uint32
	Properties map[string]string
	Type       NoteType
	Markup     Markup
	Source     *Notebook
}

func NewNote(ctx *Context, uuid uint64, title, body string) *Note {
	return &Note{context: ctx, UUID: uuid, Title: title, Body: body}
}

func (self *Note) SetBody(body string) {
	self.Body = body
	self.detectMarkup()
}

func (self *Note) detectMarkup() {
	self.Markup = MarkupNone
}

func (self *Note) mapping() map[string]interface{} {
	return map[string]interface{}{
		"UUID": self.UUID,
		"Title": self.Title,
		"Body": self.Body,
		"URI": self.URI,
		"MimeType": self.MimeType,
		"Type": self.Type,
	}
}

func (self *Note) ToHV() *C.SV {
	perl := self.context.interpreter.perl

	hv := C.Perl_newHV(perl)
	for key, value := range self.mapping() {
		cKey := C.CString(key)
		defer C.free(unsafe.Pointer(cKey))

		switch v := value.(type) {
		case int:
			C.Perl_hv_store(perl, hv, cKey, C.I32(C.strlen(cKey)),
				C.Perl_newSViv(perl, C.I64(v)), 0)
		case uint64:
			C.Perl_hv_store(perl, hv, cKey, C.I32(C.strlen(cKey)),
				C.Perl_newSViv(perl, C.I64(v)), 0) // FIXME: signedness
		case string:
			cValue := C.CString(v)
			defer C.free(unsafe.Pointer(cValue))
			valueSV := C.Perl_newSVpvn(perl, cValue, C.strlen(cValue))
			C.Perl_hv_store(perl, hv, cKey, C.I32(C.strlen(cKey)), valueSV, 0)
		default:
			continue
		}
	}
	return C.Perl_newRV_noinc(perl, (*C.SV)(unsafe.Pointer(hv)))
}
