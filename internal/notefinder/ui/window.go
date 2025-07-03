package ui

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	ra "github.com/go-shiori/go-readability"

	"notefinder/internal/notefinder/common"
	"notefinder/internal/notefinder/types"
	"notefinder/internal/notefinder/util"
)

var uriError = errors.New("Unsupported URI")

type Store interface {
	GetNotebooks() map[string]*types.Notebook
	QueryStream(query *types.Query, out chan<- *types.Note)
	Query(query *types.Query) []*types.Note
}

type Context interface {
	CloseBus()
	WriteToBus(*types.Note)
	GetRequests() chan common.Request
	ReadRequest() common.Request
	WriteRequest(common.Request)
	Refresh()
}

type Window struct {
	fyne.Window

	context          Context
	store            Store
	mx               sync.Mutex
	tabs             *container.AppTabs
	list             *widget.List
	searchInput      *widget.Entry
	statusBar        *widget.Label
	app              fyne.App
	query            *types.Query
	notebook         *types.Notebook
	selectedNote     *types.Note
	selectedListID   int
	filterByNotebook bool
	matchCase        bool
	listItemIDToNote map[widget.ListItemID]*types.Note
}

func NewWindow(ctx Context, store Store, appl fyne.App) *Window {
	mainWindow := appl.NewWindow(common.AppName)
	w := &Window{
		Window:           mainWindow,
		context:          ctx,
		store:            store,
		statusBar:        widget.NewLabel(""),
		app:              appl,
		listItemIDToNote: make(map[widget.ListItemID]*types.Note),
		query:            &types.Query{Needle: ""},
	}

	w.SetCloseIntercept(func() {
		w.context.WriteRequest(common.RequestStop)
		w.Close()
	})

	w.selectedListID = -1
	return w
}

func (w *Window) Query() *types.Query {
	return w.query
}

func (w *Window) SetQuery(query *types.Query) {
	w.query = query
	w.searchInput.SetText(query.Needle)
}

func (w *Window) Refresh() {
	w.mx.Lock()
	defer w.mx.Unlock()
	currentNotebook := w.CurrentWorkingNotebook()
	if currentNotebook != nil && w.filterByNotebook {
		w.query.Haystack = currentNotebook
	} else {
		w.query.Haystack = nil
	}
	w.query.MatchCase = w.matchCase

	ch := make(chan *types.Note)
	var mu sync.Mutex
	var notes []*types.Note

	go w.store.QueryStream(w.query, ch)

	nResults := 0
	go func() {
		for note := range ch {
			nResults++
			n := note
			mu.Lock()
			notes = append(notes, n)
			sort.Slice(notes, func(i, j int) bool { return notes[i].UUID < notes[j].UUID })
			mu.Unlock()
			status := fmt.Sprintf("%d results", nResults)
			fyne.Do(func() {
				w.statusBar.SetText(status)
				w.list.Refresh()
			})
		}
	}()

	w.list.Length = func() int {
		mu.Lock()
		defer mu.Unlock()
		sz := len(notes)
		return sz
	}
	w.list.UpdateItem = func(i widget.ListItemID, o fyne.CanvasObject) {
		mu.Lock()
		defer mu.Unlock()

		item := o.(*ClickableItem)
		item.ID = i
		item.OnTapped = openNote
		vbox := item.content.(*fyne.Container)
		rows := vbox.Objects

		topRow := rows[0].(*fyne.Container)
		detail := rows[1].(*canvas.Text)

		icon := topRow.Objects[0].(*widget.Icon)
		title := topRow.Objects[1].(*widget.Label)

		note := notes[i]

		title.TextStyle.Bold = (i == w.selectedListID)
		icon.SetResource(noteIcon(note))
		fyne.Do(func() {
			title.SetText(note.Title)
		})
		matchesText := fmt.Sprintf(" (matches:  %s)", strings.Join(note.MatchingFields, ", "))

		if note.Body != "" {
			detail.Text = util.ShortText(note.Body, 48)
		} else {
			switch note.Type {
			case types.NoteTypeBookmark:
				detail.Text = note.URI
			case types.NoteTypeFile:
				detail.Text = note.MimeType
			default:
				detail.Text = ""
			}
		}

		if w.query.Needle != "" {
			detail.Text += matchesText
		}

		detail.Refresh()
		w.listItemIDToNote[i] = note

	}

	fyne.Do(func() {
		w.list.Refresh()
	})
}

func (w *Window) ClipboardContent() string {
	return w.Clipboard().Content()
}

func (w *Window) RequestRefresh() {
	w.context.WriteRequest(common.RequestLoadData)
}

func (w *Window) Show() {
	w.SetContent(w.makeLayout())
	w.SetMaster()
	w.Resize(fyne.NewSize(800, 600))
	w.CenterOnScreen()

	ctrlQ := &desktop.CustomShortcut{KeyName: fyne.KeyQ, Modifier: fyne.KeyModifierControl}
	w.Canvas().AddShortcut(ctrlQ, func(sc fyne.Shortcut) {
		w.app.Quit()
	})

	w.RequestRefresh()
	w.ShowAndRun()
}

func (w *Window) CurrentWorkingNotebook() *types.Notebook {
	return w.notebook
}

func openNote(parent *Window, id int) {
	var note *types.Note
	var ok bool
	if id > 0 {
		note, ok = parent.listItemIDToNote[id]
		if !ok {
			return
		}
	} else {
		note = &types.Note{}
	}

	if note.URI != "" {
		if strings.HasPrefix(note.URI, "https://") ||
			strings.HasPrefix(note.URI, "http://") ||
			strings.HasPrefix(note.URI, "file://") {
			parsed, err := url.Parse(note.URI)
			if err == nil {
				go func() {
					article, err := ra.FromURL(note.URI, 30*time.Second)
					fmt.Println(article)
					if err == nil {
						log.Println(article.TextContent)
					} else {
						log.Println(err)
					}
				}()
				fyne.CurrentApp().OpenURL(parsed)
				return
			}
			dialog.ShowError(uriError, parent)
			return
		} else {
			dialog.ShowError(uriError, parent)
			return
		}
	}

	ti := NewEditorTabItem(note, parent)
	parent.tabs.Append(ti.tabItem)
	parent.tabs.Select(ti.tabItem)
}

func (w *Window) makeLayout() *fyne.Container {
	w.searchInput = w.makeSearchInput()
	tb := container.New(layout.NewFormLayout(), makeToolbar(w), w.searchInput)

	notebooks := w.store.GetNotebooks()
	names := make([]string, 0, len(notebooks))
	for name, _ := range notebooks {
		names = append(names, name)
	}

	selector := widget.NewSelect(names, func(value string) {
		w.notebook = notebooks[value]

		if w.filterByNotebook {
			w.query.Haystack = w.notebook
			w.selectedListID = -1
			w.selectedNote = nil
			w.Refresh()
		}
	})
	selector.PlaceHolder = "Current working notebook"
	notebookSelector := container.New(layout.NewHBoxLayout(),
		w.statusBar,
		selector,
		widget.NewCheck("Filter", func(value bool) {
			w.filterByNotebook = value
			w.selectedListID = -1
			w.selectedNote = nil
			w.Refresh()
		}),
		widget.NewCheck("Match case", func(value bool) {
			w.matchCase = value
			w.Refresh()
		}),
	)

	w.list = makeList(w)

	w.tabs = container.NewAppTabs(
		container.NewTabItemWithIcon("", theme.HomeIcon(), w.list),
	)

	return container.NewBorder(
		tb,
		notebookSelector,
		nil,
		nil,
		w.tabs)
}

func (w *Window) makeSearchInput() *widget.Entry {
	input := widget.NewEntry()
	input.SetPlaceHolder("Enter search query...")
	input.ActionItem = widget.NewIcon(theme.SearchIcon())

	var debounceMu sync.Mutex
	var debounceTimer *time.Timer
	const debounceDelay = 500 * time.Millisecond

	input.OnChanged = func(query string) {
		debounceMu.Lock()
		defer debounceMu.Unlock()

		if debounceTimer != nil {
			debounceTimer.Stop()
		}

		debounceTimer = time.AfterFunc(debounceDelay, func() {
			w.query = &types.Query{Needle: query}
			w.selectedListID = -1
			w.Refresh()
		},
		)

	}
	return input
}

func noteIcon(note *types.Note) fyne.Resource {
	switch note.Type {
	case types.NoteTypeBookmark:
		return theme.HistoryIcon()
	case types.NoteTypeFile:
		return theme.FileIcon()
	default:
		return theme.DocumentIcon()
	}
}
