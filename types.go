package main

import (
	"sort"
	"strings"
	"time"
	"unsafe"
)

type Implementation interface {
	LoadData() (map[uint64]*Note, error)
	PutData(*Note) error
	UpdateData(*Note, *Note) error
	DeleteData(*Note) error
	CanWrite() (bool, error)
}

type Store struct {
	data map[NoteKey]*Note
}

func NewStore() *Store {
	return &Store{data: make(map[NoteKey]*Note)}
}

func (self *Store) Get(key NoteKey) (*Note, bool) {
	v, ok := self.data[key]
	return v, ok
}

func (self *Store) Put(key NoteKey, note *Note) {
	self.data[key] = note
}

func (self *Store) Query(query *Query) []*Note {
	res := make([]*Note, 0, len(self.data))
	keys := make([]NoteKey, 0, len(self.data))

	for key, note := range self.data {
		if query.Haystack != nil && query.Haystack != key.Notebook {
			continue
		}
		if !strings.Contains(note.Title, query.Needle) {
			continue
		}
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Notebook != keys[j].Notebook {
			return uintptr(unsafe.Pointer(keys[i].Notebook)) <
				uintptr(unsafe.Pointer(keys[j].Notebook))
		}
		return keys[i].UUID < keys[j].UUID
	})

	for _, key := range keys {
		res = append(res, self.data[key])
	}

	return res
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
	MimeType string
	CreatedAt  time.Time
	ModifiedAt time.Time
	flags      uint8
	Properties map[string]string
	Type       NoteType
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
