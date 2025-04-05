package background

import (
	"fmt"
	"log"
	"sync"
	"time"

	"notefinder/internal/notefinder/common"
	"notefinder/internal/notefinder/implementation" // FIXME: shouldn't be here at all
	"notefinder/internal/notefinder/types"
)

type Manager interface {
	CloseBus()
	WriteToBus(*types.Note)
	GetRequests() chan common.Request
	ReadRequest() common.Request
	Refresh()
}

type Store interface {
	CreateNotebook(string, *types.Notebook)
	GetNotebooks() map[string]*types.Notebook
	Put(types.NoteKey, *types.Note)
	Get(types.NoteKey) (*types.Note, bool)
	Delete(types.NoteKey)
	Query(*types.Query) []*types.Note
}

type Worker struct {
	manager Manager
	store   Store
	mx      sync.Mutex
}

func NewWorker(manager Manager, store Store) *Worker {
	return &Worker{manager: manager, store: store}
}

func (w *Worker) Run() {
	defer w.manager.CloseBus()

	// FIXME: rework auto-configuration mechanism
	for name, bookmarkFile := range implementation.GetMozillaFiles() {
		bookmarkConfig := map[string]string{"path": bookmarkFile}
		w.store.CreateNotebook(name, types.NewNotebook(name,
			implementation.NewMozillaImplementation(bookmarkConfig),
			bookmarkConfig, types.NotebookAutoDiscovered))
	}

	ticker := time.NewTicker(10 * time.Second)
	doWork := func() {
		w.mx.Lock()
		defer w.mx.Unlock()
		var wg sync.WaitGroup
		notebooks := w.store.GetNotebooks()
		wg.Add(len(notebooks))
		for _, notebook := range notebooks {
			go func() {
				defer wg.Done()

				var haveUpdates bool
				data, err := notebook.LoadData()
				if err != nil {
					log.Println(err)
					return
				}

				for _, oldItem := range w.store.Query(&types.Query{Haystack: notebook}) {
					_, stillHave := data[oldItem.UUID]
					if !stillHave {
						haveUpdates = true
						w.store.Delete(types.NoteKey{Notebook: notebook, UUID: oldItem.UUID})
					}
				}

				for uuid, item := range data {
					key := types.NoteKey{Notebook: notebook, UUID: uuid}
					existingItem, ok := w.store.Get(key)
					if ok && item.SameAs(existingItem) {
						continue
					}
					item.Source = notebook

					w.store.Put(key, item)
					w.manager.WriteToBus(item)
					haveUpdates = true
				}
				if haveUpdates {
					w.manager.Refresh() // FIXME: check if this is thread-safe at all
				}
			}()
		}
		wg.Wait()
	}

	for {
		select {
		case <-ticker.C:
			doWork()
		case req := <-w.manager.GetRequests():
			switch req {
			case common.RequestLoadData:
				doWork()
			case common.RequestStop:
				fmt.Println("received graceful shutdown request")
				return
			}
		}
	}
}
