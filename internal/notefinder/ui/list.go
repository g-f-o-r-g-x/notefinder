package ui

import (
	"time"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"notefinder/internal/notefinder/types"
)

type ClickableItem struct {
	widget.BaseWidget
	content  fyne.CanvasObject
	parent   *Window
	ID       int
	OnTapped func(parent *Window, id int)
	lastTap  time.Time
}

func NewClickableItem(id int, content fyne.CanvasObject, parent *Window,
	onTapped func(parent *Window, id int)) *ClickableItem {
	ci := &ClickableItem{
		content:  content,
		parent:   parent,
		ID:       id,
		OnTapped: onTapped,
	}
	ci.ExtendBaseWidget(ci)
	return ci
}

func (c *ClickableItem) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.content)
}

func (c *ClickableItem) Tapped(_ *fyne.PointEvent) {
	c.parent.selectedListID = c.ID
	c.parent.selectedNote = c.parent.listItemIDToNote[c.ID]
	now := time.Now()
	if now.Sub(c.lastTap) < 300*time.Millisecond {
		if c.OnTapped != nil {
			c.OnTapped(c.parent, c.ID)
		}
	}
	c.lastTap = now

	c.parent.list.Refresh()
}

func makeList(win *Window) *widget.List {
	data := []*types.Note{}
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

			vbox := container.New(layout.NewVBoxLayout(), topRow, detail)
			return NewClickableItem(0, vbox, win, nil)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
		})

	return list
}
