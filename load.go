package main

import (
	"reflect"
	"time"
)

func load(ctx *Context) {
	config := ctx.Config
	ctx.Notebooks = map[string]*Notebook{
		"default": NewNotebook("default", NewFileImplementation(config), config,
			NotebookConfigured),
	}

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
				panic("Cannot load notebook data")
			}

			for uuid, item := range data {
				key := NoteKey{Notebook: notebook, UUID: uuid}
				existingItem, ok := ctx.Data.Get(key)
				if ok && reflect.DeepEqual(item, existingItem) {
					continue
				}

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
