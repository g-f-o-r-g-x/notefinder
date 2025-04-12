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

type Implementation interface {
	LoadData() (map[uint64]*Note, error)
	PutData(*Note) error
	UpdateData(*Note, *Note) error
	DeleteData(*Note) error
	CanWrite() (bool, error)
}

type NotebookType int

const (
	NotebookConfigured NotebookType = iota
	NotebookAutoDiscovered
)

type NoteKey struct {
	Notebook *Notebook
	UUID     uint64
}

type Query struct {
	Needle   string
	Haystack *Notebook
}

type Notebook struct {
	Name           string
	Config         map[string]string
	Data           map[uint64]*Note
	implementation Implementation
	Type           NotebookType
}

const (
	FlagArchived = 1 << 0
	FlagReadOnly = 1 << 1
	FlagNotify   = 1 << 2
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
	Source     *Notebook
}

func NewNote(ctx *Context, uuid uint64, title, body string) *Note {
	return &Note{context: ctx, UUID: uuid, Title: title, Body: body}
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

func NewNotebook(name string, impl Implementation, config map[string]string,
	type_ NotebookType) *Notebook {
	return &Notebook{Name: name, implementation: impl, Config: config,
		Data: make(map[uint64]*Note), Type: type_}
}

func (self *Notebook) LoadData() (map[uint64]*Note, error) {
	data, err := self.implementation.LoadData()
	return data, err
}

func (self *Notebook) CanWrite() (bool, error) {
	return self.implementation.CanWrite()
}

func (self *Notebook) PutData(note *Note) error {
	return self.implementation.PutData(note)
}

func (self *Notebook) UpdateData(oldNote *Note, newNote *Note) error {
	return self.implementation.UpdateData(oldNote, newNote)
}

func (self *Notebook) DeleteData(note *Note) error {
	return self.implementation.DeleteData(note)
}
