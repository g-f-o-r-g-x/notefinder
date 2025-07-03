package app

import (
	"strings"
	"sync"

	"notefinder/internal/notefinder/types"
	"notefinder/internal/notefinder/util"
)

type Store struct {
	context   *Context
	notebooks map[string]*types.Notebook
	data      map[types.NoteKey]*types.Note
	mx        sync.RWMutex
}

func NewStore(ctx *Context) *Store {
	return &Store{context: ctx, notebooks: readConfig(ctx),
		data: make(map[types.NoteKey]*types.Note)}
}

func (self *Store) Get(key types.NoteKey) (*types.Note, bool) {
	self.mx.RLock()
	defer self.mx.RUnlock()
	v, ok := self.data[key]
	return v, ok
}

func (self *Store) Put(key types.NoteKey, note *types.Note) {
	self.mx.Lock()
	defer self.mx.Unlock()
	//eventOnLoaded(note)
	self.data[key] = note
}

func (self *Store) Delete(key types.NoteKey) {
	self.mx.Lock()
	defer self.mx.Unlock()
	delete(self.data, key)
}

func (self *Store) QueryStream(query *types.Query, out chan<- *types.Note) {
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
		for key, desc := range note.Mapping() {
			if !desc.Searchable {
				continue
			}
			value := desc.Ptr.(*string)
			var haystack, needle string
			if !query.MatchCase {
				haystack = strings.ToLower(*value)
				needle = strings.ToLower(query.Needle)
			} else {
				haystack = *value
				needle = query.Needle
			}
			if strings.Contains(haystack, needle) {
				note.MatchingFields = append(note.MatchingFields, key)

				if !matchFound {
					out <- note
					matchFound = true
				}
			}
		}

		if note.MimeType == "application/pdf" {
			wg.Add(1)
			go func(note *types.Note) {
				defer wg.Done()
				// FIXME: memory leak!
				if util.PdfMatchesPattern(note.URI, query.Needle) {
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

func (self *Store) Query(query *types.Query) []*types.Note {
	self.mx.RLock()
	defer self.mx.RUnlock()
	res := make([]*types.Note, 0, len(self.data))

	for key, note := range self.data {
		if query.Haystack != nil && query.Haystack != key.Notebook {
			continue
		}

		res = append(res, note)
	}
	return res
}

func (self *Store) CreateNotebook(name string, notebook *types.Notebook) {
	self.notebooks[name] = notebook
}

func (self *Store) GetNotebooks() map[string]*types.Notebook {
	return self.notebooks
}
