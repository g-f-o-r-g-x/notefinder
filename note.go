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
	flags      uint8
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

func (self *Note) ToHV() *C.SV {
	perl := self.context.interpreter.perl

	hv := C.Perl_newHV(perl)
	title_key := C.CString("Title")
	title_value := C.CString(self.Title)
	defer C.free(unsafe.Pointer(title_key))
	defer C.free(unsafe.Pointer(title_value))

	body_key := C.CString("Body")
	body_value := C.CString(self.Body)
	defer C.free(unsafe.Pointer(body_key))
	defer C.free(unsafe.Pointer(body_value))

	title_sv := C.Perl_newSVpvn(perl, title_value, C.strlen(title_value))
	body_sv := C.Perl_newSVpvn(perl, body_value, C.strlen(body_value))

	C.Perl_hv_store(perl, hv, title_key, C.I32(C.strlen(title_key)), title_sv, 0)
	C.Perl_hv_store(perl, hv, body_key, C.I32(C.strlen(body_key)), body_sv, 0)

	sv := C.Perl_newRV_noinc(perl, (*C.SV)(unsafe.Pointer(hv)))
	return sv
}
