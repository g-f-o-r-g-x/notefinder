package ui

import (
	"errors"
	"strings"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"notefinder/internal/notefinder/types"
	"notefinder/internal/notefinder/util"
)

type EditorTabItem struct {
	note    *types.Note
	tabItem *container.TabItem
	viewer  *widget.RichText
	editor  *widget.Entry
	parent  *Window
}

func NewEditorTabItem(note *types.Note, parent *Window) *EditorTabItem {
	ti := &EditorTabItem{note: note, parent: parent}
	ti.viewer = widget.NewRichTextFromMarkdown(note.Body)
	ti.viewer.Wrapping = fyne.TextWrapWord
	ti.editor = widget.NewEntry()
	ti.editor.MultiLine = true
	ti.editor.Wrapping = fyne.TextWrapWord
	ti.editor.SetText(note.Body)

	if note.UUID != 0 {
		ti.editor.Hide()
	} else {
		ti.viewer.Hide()
	}

	tb := widget.NewToolbar(
		widget.NewToolbarAction(
			theme.DocumentCreateIcon(),
			func() {
				if ti.viewer.Visible() {
					ti.viewer.Hide()
					ti.editor.Show()
				} else {
					ti.viewer.Show()
					ti.editor.Hide()
				}
			},
		),
		widget.NewToolbarAction(theme.DocumentSaveIcon(),
			func() {
				entry := widget.NewEntry()
				var proposedTitle string
				if note.Title == "" && ti.editor.Text != "" {
					lines := strings.SplitN(ti.editor.Text, "\n", 1)
					if len(lines) > 0 {
						proposedTitle = util.ShortText(lines[0], 32)
						entry.SetText(proposedTitle)
					}
				} else {
					entry.SetText(note.Title)
				}

				form := dialog.NewForm("Enter title", "OK", "Cancel",
					[]*widget.FormItem{
						&widget.FormItem{Text: "", Widget: entry},
					}, func(ok bool) {
						if !ok || entry.Text == "" {
							return
						}
						note.Title = entry.Text
						parent.tabs.Selected().Text = note.Title
						parent.tabs.Refresh()

						if nb := ti.parent.CurrentWorkingNotebook(); nb != nil {
							note.Set("Body", ti.editor.Text, true)
							_ = nb.PutData(note)
							ti.parent.RequestRefresh()
						} else {
							dialog.ShowError(errors.New("Please select notebook"), parent)
							return
						}

						/* TODO:
						1. Force Worker to LoadData()
						2. Perform SameAs on item
						3. On errors save note to drafts
						*/

					}, parent)
				form.Show()
				parent.Canvas().Focus(entry)
			},
		),
	)

	togglableView := container.New(layout.NewStackLayout(), ti.viewer, ti.editor)
	tabContent := container.NewBorder(tb, nil, nil, nil, togglableView)
	ti.tabItem = container.NewTabItemWithIcon(note.Title, noteIcon(note), tabContent)
	/*
		parent.tabs.Append(tabItem)
		parent.tabs.Select(tabItem)
	*/
	return ti
}
