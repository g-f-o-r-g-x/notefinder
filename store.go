package main

import (
	"fmt"
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
}

func NewStore(ctx *Context) *Store {
	return &Store{context: ctx, data: make(map[NoteKey]*Note)}
}

func (self *Store) Get(key NoteKey) (*Note, bool) {
	v, ok := self.data[key]
	return v, ok
}

func (self *Store) Put(key NoteKey, note *Note) {
	self.data[key] = note
}

func (self *Store) Delete(key NoteKey) {
	delete(self.data, key)
}

func sendNote(out chan<- *Note, note *Note, query *Query, nResults *int) {
	fmt.Printf("Sending note titled '%s' for query: '%s'\n", note.Title, query.Needle)
	*nResults++
	out <- note
}

func (self *Store) QueryStream(query *Query, out chan<- *Note) {
	//	var wg sync.WaitGroup
	nResults := 0

	for key, note := range self.data {
		if query.Haystack != nil && query.Haystack != key.Notebook {
			continue
		}

		if query.Needle == "" || strings.Contains(note.Title, query.Needle) {
			sendNote(out, note, query, &nResults)
			continue
		}
		/*
			if note.MimeType == "application/pdf" {
				wg.Add(1)
				go func(note *Note) {
					defer wg.Done()
					pdfFilePath := strings.TrimPrefix(note.URI, "file://")
					if pdfMatchesPattern(pdfFilePath, query.Needle) {
						out <- note
					}
				}(note)
			}
		*/
	}
	/*
		go func() {
			wg.Wait()
			close(out)
		}()
	*/
	fmt.Println("Results:", nResults)
	close(out)
}

func (self *Store) Query(query *Query) []*Note {
	res := make([]*Note, 0, len(self.data))
	keys := make([]NoteKey, 0, len(self.data))

	var mx sync.Mutex

	for key, note := range self.data {
		if query.Haystack != nil && query.Haystack != key.Notebook {
			continue
		}

		fmt.Printf("query is: '%s'\n", query.Needle)
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
					fmt.Println("Match found:", note.Title)
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
