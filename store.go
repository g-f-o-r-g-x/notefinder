package main

import (
	"sort"
	"strings"
	"sync"
	"unsafe"
)

type NoteKey struct {
	Notebook *Notebook
	UUID     uint64
}

type Query struct {
	Needle   string
	Haystack *Notebook
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
				pdfFilePath := strings.TrimPrefix(note.URI, "file://")
				if pdfMatchesPattern(pdfFilePath, query.Needle) {
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
	keys := make([]NoteKey, 0, len(self.data))

	var mx sync.Mutex

	for key, note := range self.data {
		if query.Haystack != nil && query.Haystack != key.Notebook {
			continue
		}

		if query.Needle == "" {
			keys = append(keys, key)
			continue
		}

		if strings.Contains(note.Title, query.Needle) {
			keys = append(keys, key)
			continue
		}

		// FIXME: just a temporary hack to search through PDFs
		if note.MimeType == "application/pdf" {
			go func() {
				pdfFilePath := strings.TrimPrefix(note.URI, "file://")
				if pdfMatchesPattern(pdfFilePath, query.Needle) {
					mx.Lock()
					defer mx.Unlock()
					keys = append(keys, key)
				}
			}()
		}
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
