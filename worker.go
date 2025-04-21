package main

import (
	"log"
	"sync"
	"time"
)

type Worker struct {
	context  *Context
	toInterp chan<- *Note
}

func NewWorker(ctx *Context, toInterp chan<- *Note) *Worker {
	return &Worker{context: ctx, toInterp: toInterp}
}

func (w *Worker) Run() {
	defer close(w.toInterp)
	for name, bookmarkFile := range getMozillaFiles() {
		bookmarkConfig := map[string]string{"path": bookmarkFile}
		w.context.Notebooks[name] = NewNotebook(name,
			NewMozillaImplementation(w.context, bookmarkConfig),
			bookmarkConfig, NotebookAutoDiscovered)
	}

	ticker := time.NewTicker(10 * time.Second)
	doWork := func() {
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
					w.toInterp <- item
					haveUpdates = true
				}
				if haveUpdates {
					w.context.Window.Refresh()
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
				return
			}
		}
	}
}
