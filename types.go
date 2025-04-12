package main

import (
	"time"
)

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

func NewNote(uuid uint64, title, body string) *Note {
	return &Note{UUID: uuid, Title: title, Body: body}
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
