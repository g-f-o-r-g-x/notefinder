package main

import (
	"errors"
	"fmt"
	"net/url"
	"runtime"
	"strings"
	"sync"
//	"time"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var uriError = errors.New("Unsupported URI")

type Window struct {
	window           fyne.Window
	list             *widget.List
	searchInput      *widget.Entry
	app              fyne.App
	context          *Context
	query            *Query
	notebook         *Notebook
	selectedNote     *Note
	selectedListID   int
	filterByNotebook bool
	listItemIDToNote map[widget.ListItemID]*Note
}

func NewWindow(ctx *Context) *Window {
	mainWindow := ctx.Application.NewWindow(appName)
	w := &Window{window: mainWindow, app: ctx.Application, context: ctx,
		listItemIDToNote: make(map[widget.ListItemID]*Note), query: &Query{Needle: ""}}

	w.selectedListID = -1
	return w
}

func (w *Window) Query() *Query {
	return w.query
}

func (w *Window) SetQuery(query *Query) {
	w.query = query
	w.searchInput.SetText(query.Needle)
}

func openNote(parent *Window, id int) {
	note, ok := parent.listItemIDToNote[id]
	if !ok {
		return
	}

	if note.URI != "" {
		if strings.HasPrefix(note.URI, "https://") ||
			strings.HasPrefix(note.URI, "http://") ||
			strings.HasPrefix(note.URI, "file://") {
			parsed, err := url.Parse(note.URI)
			if err == nil {
				fyne.CurrentApp().OpenURL(parsed)
				return
			}
			dialog.ShowError(uriError, parent.window)
			return
		} else {
			dialog.ShowError(uriError, parent.window)
			return
		}
	}

	textViewer := widget.NewRichTextFromMarkdown(note.Body)
	textViewer.Wrapping = fyne.TextWrapWord
	textEditor := widget.NewEntry()
	textEditor.MultiLine = true
	textEditor.Wrapping = fyne.TextWrapWord
	textEditor.SetText(note.Body)
	textEditor.Hide()

	tb := widget.NewToolbar(
		widget.NewToolbarAction(
			theme.DocumentCreateIcon(),
			func() {
				textViewer.Hide()
				textEditor.Show()
			},
		),
		widget.NewToolbarAction(theme.DocumentSaveIcon(),
			func() {
				textEditor.Hide()
				textViewer.Show()
			},
		),
	)

	v := container.New(layout.NewStackLayout(), textViewer, textEditor)
	c := container.NewBorder(tb, nil, nil, nil, v)

	editorWindow := parent.app.NewWindow(note.Title)
	editorWindow.SetContent(c)
	editorWindow.CenterOnScreen()
	editorWindow.Resize(fyne.NewSize(540, 460))
	editorWindow.Show()
}

func (w *Window) Refresh() {
	fmt.Println("Entry to:", currentFunction())
	currentNotebook := w.CurrentWorkingNotebook()
	if currentNotebook != nil && w.filterByNotebook {
		w.query.Haystack = currentNotebook
	} else {
		w.query.Haystack = nil
	}

	ch := make(chan *Note)
	var mu sync.Mutex
	var notes []*Note

	go w.context.Data.QueryStream(w.query, ch)

	nResults := 0
	go func() {
		fmt.Println("len(ch) before reading", len(ch))
		for note := range ch {
			nResults++
			n := note // capture
			mu.Lock()
			notes = append(notes, n)
			mu.Unlock()

//			time.AfterFunc(10*time.Millisecond, func() {
				w.window.Canvas().Refresh(w.list)
//			})
		}
		fmt.Println("len(ch) afterwards", len(ch))
		fmt.Println("Results to display:", nResults)
		fmt.Println("w.List.Length():", w.list.Length())
		fmt.Println("-------------------")
	}()

	w.list.Length = func() int {
		fmt.Println("Entry to:", currentFunction())
		mu.Lock()
		defer mu.Unlock()
		sz := len(notes)
		//		fmt.Println("Length() returning:", sz)
		return sz
	}
	w.list.UpdateItem = func(i widget.ListItemID, o fyne.CanvasObject) {
		fmt.Println("Entry to:", currentFunction())
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
		title.SetText(note.Title)

		if note.Body != "" {
			detail.Text = shortText(note.Body, 64)
		} else {
			switch note.Type {
			case NoteTypeBookmark:
				detail.Text = note.URI
			case NoteTypeFile:
				detail.Text = note.MimeType
			default:
				detail.Text = ""
			}
		}

		detail.Refresh()
		w.listItemIDToNote[i] = note

	}

	w.list.Refresh()
}

func (w *Window) Refresh2() {
	currentNotebook := w.CurrentWorkingNotebook()
	if currentNotebook != nil && w.filterByNotebook {
		w.query.Haystack = w.CurrentWorkingNotebook()
	} else {
		w.query.Haystack = nil
	}
	data := w.context.Data.Query(w.query)

	w.list.Length = func() int {
		fmt.Println("Entry to:", currentFunction())
		return len(data)
	}
	w.list.UpdateItem = func(i widget.ListItemID, o fyne.CanvasObject) {
		fmt.Println("Entry to:", currentFunction())
		item := o.(*ClickableItem)
		item.ID = i
		item.OnTapped = openNote

		vbox := item.content.(*fyne.Container)
		rows := vbox.Objects

		topRow := rows[0].(*fyne.Container)
		detail := rows[1].(*canvas.Text)

		icon := topRow.Objects[0].(*widget.Icon)
		title := topRow.Objects[1].(*widget.Label)

		note := data[i]

		// Set title to bold if item is selected
		if i == w.selectedListID {
			title.TextStyle.Bold = true
		} else {
			title.TextStyle.Bold = false
		}

		// Update the rest of the details and text
		icon.SetResource(noteIcon(note))
		title.SetText(note.Title)

		if note.Body != "" {
			detail.Text = shortText(note.Body, 64)
		} else {
			switch note.Type {
			case NoteTypeBookmark:
				detail.Text = note.URI
			case NoteTypeFile:
				detail.Text = note.MimeType
			default:
				detail.Text = ""
			}
		}

		// Refresh the detail text after the update
		detail.Refresh()

		// Store the note in the mapping
		w.listItemIDToNote[i] = note
	}

	w.list.Refresh()
}

func (w *Window) ClipboardContent() string {
	return w.window.Clipboard().Content()
}

func (w *Window) Show() {
	w.window.SetContent(w.makeLayout())
	w.window.SetMaster()
	w.window.Resize(fyne.NewSize(800, 600))
	w.window.CenterOnScreen()

	w.context.Requests <- RequestLoadData
	w.window.ShowAndRun()
}

func (w *Window) CurrentWorkingNotebook() *Notebook {
	return w.notebook
}

func (w *Window) makeLayout() *fyne.Container {
	w.searchInput = w.context.MainWindow.makeSearchInput()
	tb := container.New(layout.NewFormLayout(), makeToolbar(w.context), w.searchInput)

	names := make([]string, 0, len(w.context.Notebooks))
	for name, _ := range w.context.Notebooks {
		names = append(names, name)
	}

	selector := widget.NewSelect(names, func(value string) {
		w.notebook = w.context.Notebooks[value]

		if w.filterByNotebook {
			w.query.Haystack = w.notebook
			w.selectedListID = -1
			w.selectedNote = nil
			w.Refresh()
		}
	})
	selector.PlaceHolder = "Current working notebook"
	notebookSelector := container.New(layout.NewHBoxLayout(),
		selector,
		widget.NewCheck("Filter", func(value bool) {
			w.filterByNotebook = value
			w.selectedListID = -1
			w.selectedNote = nil
			w.Refresh()
		}),
	)

	w.list = makeList(w.context)

	return container.NewBorder(
		tb,
		notebookSelector,
		nil,
		nil,
		w.list)
}

func currentFunction() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return "?"
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "?"
	}

	return fn.Name()
}

func (w *Window) makeSearchInput() *widget.Entry {
	input := widget.NewEntry()
	input.SetPlaceHolder("Enter search query...")
	input.ActionItem = widget.NewIcon(theme.SearchIcon())

	//	var debounceMu sync.Mutex
	//	var debounceTimer *time.Timer
	//	const debounceDelay = 500 * time.Millisecond // adjust delay as needed

	//input.OnChanged = func(query string) {
	input.OnSubmitted = func(query string) {
		//		debounceMu.Lock()
		//		defer debounceMu.Unlock()

		//		if debounceTimer != nil {
		//			debounceTimer.Stop()
		//		}

		//		debounceTimer = time.AfterFunc(debounceDelay, func() {
		w.query = &Query{Needle: query}
		w.selectedListID = -1
		w.Refresh()
	}
	//)

	//}

	return input
}

/*
func (w *Window) makeSearchInput() *widget.Entry {

		input := widget.NewEntry()
		input.SetPlaceHolder("Enter search query...")
		input.ActionItem = widget.NewIcon(theme.SearchIcon())
		input.OnChanged = func(query string) {
			w.query = &Query{Needle: query}
			w.selectedListID = -1
			w.Refresh()
		}

		return input
	}
*/
func noteIcon(note *Note) fyne.Resource {
	switch note.Type {
	case NoteTypeBookmark:
		return theme.HistoryIcon()
	case NoteTypeFile:
		return theme.FileIcon()
	default:
		return theme.DocumentIcon()
	}
}

func shortText(in string, limit int) string {
	lines := strings.Split(in, "\n")
	l := lines[0]

	if len(l) > limit {
		res := []string{}
		words := strings.Split(l, " ")
		for _, word := range words {
			if len(strings.Join(res, " ")+word) <= limit*2 {
				res = append(res, word)
			}
		}

		// Some jabberwocky so we could not even collect a word
		if len(res) == 0 {
			return l[:limit] + "..."
		}
		return strings.Join(res, " ") + "..."
	}

	return l
}
