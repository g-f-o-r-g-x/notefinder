package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Worker struct {
	context *Context
	mx      sync.Mutex
}

func NewWorker(ctx *Context) *Worker {
	return &Worker{context: ctx}
}

func (w *Worker) Run() {
	defer close(w.context.Bus)
	// FIXME: rework auto-configuration mechanism
	for name, bookmarkFile := range getMozillaFiles() {
		bookmarkConfig := map[string]string{"path": bookmarkFile}
		w.context.Notebooks[name] = NewNotebook(name,
			NewMozillaImplementation(w.context, bookmarkConfig),
			bookmarkConfig, NotebookAutoDiscovered)
	}

	ticker := time.NewTicker(10 * time.Second)
	doWork := func() {
		w.mx.Lock()
		defer w.mx.Unlock()
		var wg sync.WaitGroup
		wg.Add(len(w.context.Notebooks))
		for _, notebook := range w.context.Notebooks {
			go func() {
				defer wg.Done()

				var haveUpdates bool
				data, err := notebook.LoadData()
				if err != nil {
					log.Println(err)
					return
				}

				for _, oldItem := range w.context.Data.Query(&Query{Haystack: notebook}) {
					_, stillHave := data[oldItem.UUID]
					if !stillHave {
						haveUpdates = true
						w.context.Data.Delete(NoteKey{Notebook: notebook, UUID: oldItem.UUID})
					}
				}

				for uuid, item := range data {
					key := NoteKey{Notebook: notebook, UUID: uuid}
					existingItem, ok := w.context.Data.Get(key)
					if ok && item.SameAs(existingItem) {
						continue
					}
					item.Source = notebook

					w.context.Data.Put(key, item)
					w.context.Bus <- item
					haveUpdates = true
				}
				if haveUpdates {
					w.context.Window.Refresh() // FIXME: check if this is thread-safe at all
				}
			}()
		}
		wg.Wait()
	}

	for {
		select {
		case <-ticker.C:
			doWork()
		case req := <-w.context.Requests:
			switch req {
			case RequestLoadData:
				doWork()
			case RequestStop:
				fmt.Println("received graceful shutdown request")
				return
			}
		}
	}
}
