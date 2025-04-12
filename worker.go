package main

import (
	"log"
	"reflect"
	"time"
)

func worker(ctx *Context, toInterp chan<- *Note) {
	for name, bookmarkFile := range getMozillaFiles() {
		bookmarkConfig := map[string]string{"path": bookmarkFile}
		ctx.Notebooks[name] = NewNotebook(name,
			NewMozillaImplementation(ctx, bookmarkConfig),
			bookmarkConfig, NotebookAutoDiscovered)
	}

	ticker := time.NewTicker(30 * time.Second)
	doWork := func() {
		var haveUpdates bool
		for _, notebook := range ctx.Notebooks {
			data, err := notebook.LoadData()
			if err != nil {
				log.Println(err)
				continue
			}

			for _, oldItem := range ctx.Data.Query(&Query{Haystack: notebook}) {
				_, stillHave := data[oldItem.UUID]
				if !stillHave {
					ctx.Data.Delete(NoteKey{Notebook: notebook, UUID: oldItem.UUID})
				}
			}

			for uuid, item := range data {
				key := NoteKey{Notebook: notebook, UUID: uuid}
				existingItem, ok := ctx.Data.Get(key)
				if ok && reflect.DeepEqual(item, existingItem) {
					continue
				}
				item.Source = notebook

				ctx.Data.Put(key, item)
				toInterp <- item
				haveUpdates = true
			}
		}
		if haveUpdates {
			ctx.MainWindow.Refresh()
		}
	}

	for {
		select {
		case <-ticker.C:
			doWork()
			return
		case req := <-ctx.Requests:
			switch req {
			case RequestLoadData:
				doWork()
				return
			case RequestStop:
				return
			}
		}
	}
}
