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
	FlagStarred  = 1 << 3
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
	MatchingFields []string
	Source     *Notebook
}

func NewNote(ctx *Context, uuid uint64, title string) *Note {
	return &Note{context: ctx, UUID: uuid, Title: title,
		MatchingFields: make([]string, 0, 4)}
}

func (this *Note) SameAs(other *Note) bool {
	if other.ModifiedAt != this.ModifiedAt {
		return false
	}
	return (
		this.UUID == other.UUID &&
		this.Title == other.Title &&
		this.Body == other.Body &&
		this.URI == other.URI &&
		this.flags == other.flags)
}

func (n *Note) SetFlag(flag uint32) {
	n.flags |= flag
}

func (n *Note) UnsetFlag(flag uint32) {
	n.flags &^= flag
}

func (n *Note) FlagIsSet(flag uint32) bool {
	return n.flags&flag != 0
}

func (n *Note) FlagsString() string {
	var out [32]rune
	for i := 31; i >= 0; i-- {
		if n.flags&(1<<uint(i)) != 0 {
			out[31-i] = '+'
		} else {
			out[31-i] = '-'
		}
	}
	return string(out[:])
}

func (self *Note) Set(key string, value interface{}, act bool) {
	if_ := self.mapping()[key].Ptr

	switch ptr := if_.(type) {
	case *string:
		*ptr = value.(string)
	default:
		return
	}

	if !act {
		return
	}

	switch key {
	case "Body":
		self.detectMarkup()
	default:
		return
	}
}

func (self *Note) detectMarkup() {
	self.Markup = MarkupNone
}

type FieldDescription struct {
	Ptr        interface{}
	Searchable bool
}

func (self *Note) mapping() map[string]*FieldDescription {
	return map[string]*FieldDescription{
		"UUID":     &FieldDescription{Ptr: &self.UUID},
		"Title":    &FieldDescription{Ptr: &self.Title, Searchable: true},
		"Body":     &FieldDescription{Ptr: &self.Body, Searchable: true},
		"URI":      &FieldDescription{Ptr: &self.URI, Searchable: true},
		"MimeType": &FieldDescription{Ptr: self.MimeType},
		"Type":     &FieldDescription{Ptr: &self.Type},
		"flags":    &FieldDescription{Ptr: &self.flags},
	}
}

func (self *Note) ToHV() *C.SV {
	perl := self.context.interpreter.perl

	hv := C.Perl_newHV(perl)
	for key, desc := range self.mapping() {
		cKey := C.CString(key)
		defer C.free(unsafe.Pointer(cKey))

		value := desc.Ptr
		switch v := value.(type) {
		case *uint32:

		case *NoteType:
			C.Perl_hv_store(perl, hv, cKey, C.I32(C.strlen(cKey)),
				C.Perl_newSViv(perl, C.I64(*v)), 0)
		case *uint64:
			C.Perl_hv_store(perl, hv, cKey, C.I32(C.strlen(cKey)),
				C.Perl_newSViv(perl, C.I64(*v)), 0) // FIXME: signedness
		case *string:
			cValue := C.CString(*v)
			defer C.free(unsafe.Pointer(cValue))
			valueSV := C.Perl_newSVpvn(perl, cValue, C.strlen(cValue))
			C.Perl_hv_store(perl, hv, cKey, C.I32(C.strlen(cKey)), valueSV, 0)
		default:
			continue
		}
	}
	return C.Perl_newRV_noinc(perl, (*C.SV)(unsafe.Pointer(hv)))
}
