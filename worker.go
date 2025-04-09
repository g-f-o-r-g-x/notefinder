package main

import (
	"log"
	"reflect"
	"time"
)

func worker(ctx *Context) {
	for name, bookmarkFile := range getMozillaFiles() {
		bookmarkConfig := map[string]string{"path": bookmarkFile}
		ctx.Notebooks[name] = NewNotebook(name,
			NewMozillaImplementation(bookmarkConfig),
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
				log.Println(oldItem.UUID, oldItem.Title)

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
		case req := <-ctx.Requests:
			switch req {
			case RequestLoadData:
				doWork()
			case RequestStop:
				return
			}
		}
	}
}
