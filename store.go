package main

import (
	"strings"
	"sync"
)

type NoteKey struct {
	Notebook *Notebook
	UUID     uint64
}

type QueryMethod int

const (
	QueryWithIndex QueryMethod = iota
	QueryDirect
	QueryRegexp
	QueryComplex
)

type Query struct {
	Needle   string
	Haystack *Notebook
	Method   QueryMethod
}

type Store struct {
	context *Context
	data    map[NoteKey]*Note
	mx      sync.RWMutex
}

func NewStore(ctx *Context) *Store {
	return &Store{context: ctx, data: make(map[NoteKey]*Note)}
}

func (self *Store) Get(key NoteKey) (*Note, bool) {
	self.mx.RLock()
	defer self.mx.RUnlock()
	v, ok := self.data[key]
	return v, ok
}

func (self *Store) Put(key NoteKey, note *Note) {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.data[key] = note
}

func (self *Store) Delete(key NoteKey) {
	self.mx.Lock()
	defer self.mx.Unlock()
	delete(self.data, key)
}

func (self *Store) QueryStream(query *Query, out chan<- *Note) {
	self.mx.RLock()
	defer self.mx.RUnlock()
	var wg sync.WaitGroup

	for key, note := range self.data {
		note.MatchingFields = make([]string, 0, 4)
		if query.Haystack != nil && query.Haystack != key.Notebook {
			continue
		}

		if query.Needle == "" {
			out <- note
			continue
		}

		var matchFound bool
		for key, desc := range note.mapping() {
			if !desc.Searchable {
				continue
			}
			value := desc.Ptr.(*string)
			if strings.Contains(strings.ToLower(*value), strings.ToLower(query.Needle)) {
				note.MatchingFields = append(note.MatchingFields, key)

				if !matchFound {
					out <- note
					matchFound = true
				}
			}
		}

		if note.MimeType == "application/pdf" {
			wg.Add(1)
			go func(note *Note) {
				defer wg.Done()
				// FIXME: memory leak!
				if pdfMatchesPattern(note.URI, query.Needle) {
					note.MatchingFields = append(note.MatchingFields, "PDF content")
					out <- note
				}
			}(note)
		}
	}
	go func() {
		wg.Wait()
		close(out)
	}()
}

func (self *Store) Query(query *Query) []*Note {
	self.mx.RLock()
	defer self.mx.RUnlock()
	res := make([]*Note, 0, len(self.data))

	for key, note := range self.data {
		if query.Haystack != nil && query.Haystack != key.Notebook {
			continue
		}

		res = append(res, note)
	}
	return res
}
