package main

import (
	"errors"
	"net/url"
	"strings"

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
	statusBar        *widget.Label
	app              fyne.App
	context          *Context
	query            *Query
	notebook         *Notebook
	selectedNote     map[*Note]int // n clicks
	filterByNotebook bool
	ListItemIDToNote map[widget.ListItemID]*Note
}

func NewWindow(ctx *Context) *Window {
	mainWindow := ctx.Application.NewWindow(appName)
	w := &Window{window: mainWindow, app: ctx.Application, context: ctx,
		ListItemIDToNote: make(map[widget.ListItemID]*Note), query: &Query{Needle: ""}}

	return w
}

func (w *Window) Query() *Query {
	return w.query
}

func (w *Window) SetQuery(query *Query) {
	w.query = query
	w.searchInput.SetText(query.Needle)
}

func (w *Window) Refresh() {
	w.statusBar.Show()
	w.statusBar.SetText("Refreshing...")

	currentNotebook := w.CurrentWorkingNotebook()
	if currentNotebook != nil && w.filterByNotebook {
		w.query.Haystack = w.CurrentWorkingNotebook()
	} else {
		w.query.Haystack = nil
	}
	data := w.context.Data.Query(w.query)

	w.list.Length = func() int {
		return len(data)
	}
	w.list.UpdateItem = func(i widget.ListItemID, o fyne.CanvasObject) {
		item := o.(*ClickableItem)
		item.ID = i
		item.OnTapped = func(id int) {
			note, ok := w.ListItemIDToNote[id]
			if note.URI != "" {
				if strings.HasPrefix(note.URI, "https://") ||
					strings.HasPrefix(note.URI, "http://") ||
					strings.HasPrefix(note.URI, "file://") {
					parsed, err := url.Parse(note.URI)
					if err == nil {
						fyne.CurrentApp().OpenURL(parsed)
						return
					}
					dialog.ShowError(uriError, w.window)
				} else {
					dialog.ShowError(uriError, w.window)
					return
				}
			}
			if !ok {
				return
			}
			textViewer := widget.NewRichTextWithText(note.Body)
			textViewer.Wrapping = fyne.TextWrapWord
			textEditor := widget.NewEntry()
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

			editorWindow := w.app.NewWindow(note.Title)
			editorWindow.SetContent(c)
			editorWindow.CenterOnScreen()
			editorWindow.Resize(fyne.NewSize(540, 460))
			editorWindow.Show()
		}

		vbox := item.content.(*fyne.Container)
		rows := vbox.Objects

		topRow := rows[0].(*fyne.Container)
		detail := rows[1].(*canvas.Text)

		icon := topRow.Objects[0].(*widget.Icon)
		title := topRow.Objects[1].(*widget.Label)

		note := data[i]

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
		w.context.MainWindow.ListItemIDToNote[i] = note
	}
	w.list.Refresh()
	w.statusBar.Hide()
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
			w.Refresh()
		}
	})
	selector.PlaceHolder = "Current working notebook"
	notebookSelector := container.New(layout.NewHBoxLayout(),
		selector,
		widget.NewCheck("Filter", func(value bool) {
			w.filterByNotebook = value
			w.Refresh()
		}),
	)

	w.list = makeList(w.context)

	w.statusBar = widget.NewLabel("")
	w.statusBar.Hide()

	return container.NewBorder(
		tb,
		notebookSelector,
		nil,
		nil,
		w.list)
}

func (w *Window) makeSearchInput() *widget.Entry {
	input := widget.NewEntry()
	input.SetPlaceHolder("Enter search query...")
	input.ActionItem = widget.NewIcon(theme.SearchIcon())
	input.OnChanged = func(query string) {
		w.query = &Query{Needle: query}
		w.Refresh()
	}

	return input
}

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
