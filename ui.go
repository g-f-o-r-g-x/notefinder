package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Window struct {
	window           fyne.Window
	list             *widget.List
	app              fyne.App
	context          *Context
	query            string
	ListItemIDToNote map[widget.ListItemID]*Note
	// TODO: add search input and status bar
}

func NewWindow(ctx *Context) *Window {
	mainWindow := ctx.Application.NewWindow(appName)
	w := &Window{window: mainWindow, app: ctx.Application, context: ctx,
		ListItemIDToNote: make(map[widget.ListItemID]*Note)}

	return w
}

func (w *Window) Query() string {
	return w.query
}

func (w *Window) SetQuery(query string) {
	w.query = query
}

func (w *Window) makeLayout() *fyne.Container {
	w.list = makeList(w.context)
	return container.NewBorder(
		container.New(layout.NewFormLayout(), makeToolbar(w.context),
			w.context.MainWindow.makeSearchInput()),
		widget.NewLabel(""),
		nil,
		nil,
		w.list)
}

func noteIcon(note *Note) fyne.Resource {
	switch note.Type {
	case NoteTypeBookmark:
		return theme.HistoryIcon()
	default:
		return theme.DocumentIcon()
	}
}

func (w *Window) Refresh() {
	data := w.context.Data.Query(w.query)
	w.list.Length = func() int {
		return len(data)
	}
	w.list.UpdateItem = func(i widget.ListItemID, o fyne.CanvasObject) {
		rows := o.(*fyne.Container).Objects
		topRow := rows[0].(*fyne.Container)
		detail := rows[1].(*canvas.Text)

		icon := topRow.Objects[0].(*widget.Icon)
		title := topRow.Objects[1].(*widget.Label)

		icon.SetResource(noteIcon(data[i]))
		title.SetText(data[i].Title)
		detail.Text = shortText(data[i].Body, 40)
		w.context.MainWindow.ListItemIDToNote[i] = data[i]
		detail.Refresh()
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

		// Some jabberwocky so we could not even collect one word for title
		if len(res) == 0 {
			return l[:limit] + "..."
		}
		return strings.Join(res, " ") + "..."
	}

	return l
}

func makeToolbar(ctx *Context) *widget.Toolbar {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.HomeIcon(), func() {
			ctx.MainWindow.SetQuery("")
			ctx.MainWindow.Refresh()
		}),
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {}),
		widget.NewToolbarAction(theme.MediaRecordIcon(), func() {}),
		widget.NewToolbarAction(theme.VisibilityOffIcon(), func() {}),
		widget.NewToolbarAction(theme.DeleteIcon(), func() {}),
		widget.NewToolbarAction(theme.ContentPasteIcon(), func() {
			content := ctx.MainWindow.ClipboardContent()
			note := NewNote(0, shortText(content, 32), content+"\n")
			if err := ctx.Notebooks[0].PutData(note); err != nil {
				log.Println(err)
			}
			log.Println(content)
		}),
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			ctx.Requests <- RequestLoadData
		}),
		widget.NewToolbarAction(theme.InfoIcon(), func() {
			img := canvas.NewImageFromResource(appLogo)
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(128, 128))

			headerText := fmt.Sprintf("%s %.1f", appName, appVersion)
			header := container.New(layout.NewCenterLayout(),
				widget.NewLabelWithStyle(headerText, fyne.TextAlignCenter,
					fyne.TextStyle{Bold: true}))

			authorLabel := widget.NewLabel("Author: Sergey S.")
			licenseLink := widget.NewHyperlink("License", &url.URL{
				Scheme: "https",
				Host:   "opensource.org",
				Path:   "/license/bsd-3-clause",
			})

			footer := container.NewVBox(
				container.New(layout.NewCenterLayout(), authorLabel),
				container.New(layout.NewCenterLayout(), licenseLink),
			)

			content := container.NewBorder(header, footer, nil, nil, img)

			dialog.ShowCustom("About", "Close", content, ctx.MainWindow.window)

		}),
	)
}

func (w *Window) makeSearchInput() *widget.Entry {
	input := widget.NewEntry()
	input.SetPlaceHolder("Enter search query...")
	input.OnChanged = func(query string) {
		w.query = query
		w.Refresh()
	}

	return input
}

func makeList(ctx *Context) *widget.List {
	data := []*Note{}
	list := widget.NewList(
		func() int {
			return len(data)
		},
		func() fyne.CanvasObject {
			icon := widget.NewIcon(theme.ConfirmIcon())
			title := widget.NewLabel("Title")
			topRow := container.New(layout.NewHBoxLayout(), icon, title)

			detail := canvas.NewText("Brief content", theme.ForegroundColor())
			detail.TextStyle.Italic = true

			return container.New(layout.NewVBoxLayout(), topRow, detail)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			rows := o.(*fyne.Container).Objects
			topRow := rows[0].(*fyne.Container)
			detail := rows[1].(*canvas.Text)

			icon := topRow.Objects[0].(*widget.Icon)
			title := topRow.Objects[1].(*widget.Label)

			icon.SetResource(theme.ConfirmIcon())
			title.SetText(data[i].Title)
			detail.Text = shortText(data[i].Body, 64)
			detail.Refresh()

			ctx.MainWindow.ListItemIDToNote[i] = data[i]
		})

	list.OnSelected = func(id widget.ListItemID) {
		go func() {
			log.Println("entry to OnSelected")
			time.Sleep(300 * time.Millisecond)
			note, ok := ctx.MainWindow.ListItemIDToNote[id]
			if !ok {
				log.Println("nothing :(")
				return
			}
			e := widget.NewRichTextWithText(note.Body)
			e.Wrapping = fyne.TextWrapWord
			e2 := widget.NewEntry()
			fmt.Println(ctx.MainWindow.ListItemIDToNote)
			e2.SetText(note.Body)
			e2.Hide()

			tb := widget.NewToolbar(widget.NewToolbarAction(
				theme.DocumentCreateIcon(),
				func() {
					e.Hide()
					e2.Show()
					e2.Refresh()
				}))

			v := container.New(layout.NewVBoxLayout(), e, e2)
			c := container.NewBorder(tb, nil, nil, nil, v)

			w := ctx.Application.NewWindow(note.Title)
			w.SetContent(c)
			w.CenterOnScreen()
			e.Resize(fyne.NewSize(540, 460))
			w.Resize(fyne.NewSize(540, 460))
			w.Show()

		}()
	}

	return list
}
