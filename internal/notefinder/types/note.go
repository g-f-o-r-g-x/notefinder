package types

import (
	_ "golang.org/x/crypto/chacha20poly1305"
	"time"
)

const (
	FlagArchived  = 1 << 0
	FlagReadOnly  = 1 << 1
	FlagNotify    = 1 << 2
	FlagStarred   = 1 << 3
	FlagEncrypted = 1 << 4
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
	NoteTypeTodoList
)

type Note struct {
	Source               *Notebook
	UUID                 uint64
	Title                string
	Body                 string
	Tags                 []string
	URI                  string
	MimeType             string
	CreatedAt            time.Time
	ModifiedAt           time.Time
	flags                uint32
	Type                 NoteType
	Markup               Markup
	LastMatchingQuery    *Query
	MatchingFields       []string
	AdditionalProperties map[string]string
}

func NewNote(uuid uint64, title string) *Note {
	return &Note{UUID: uuid, Title: title,
		MatchingFields: make([]string, 0, 4),
		Tags:           make([]string, 0, 0)}
}

func (this *Note) SameAs(other *Note) bool {
	if other.ModifiedAt != this.ModifiedAt {
		return false
	}
	return (this.UUID == other.UUID &&
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
	if_ := self.Mapping()[key].Ptr

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

func (self *Note) Mapping() map[string]*FieldDescription {
	return map[string]*FieldDescription{
		"UUID":     &FieldDescription{Ptr: &self.UUID},
		"Title":    &FieldDescription{Ptr: &self.Title, Searchable: true},
		"Body":     &FieldDescription{Ptr: &self.Body, Searchable: true},
		"URI":      &FieldDescription{Ptr: &self.URI},
		"MimeType": &FieldDescription{Ptr: &self.MimeType},
		"Type":     &FieldDescription{Ptr: &self.Type},
		"flags":    &FieldDescription{Ptr: &self.flags},
	}
}
