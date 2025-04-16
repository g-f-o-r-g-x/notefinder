package main

import (
	"errors"
	"fmt"
	"log"
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

func makeToolbar(ctx *Context) *widget.Toolbar {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.HomeIcon(), func() {
			ctx.Window.SetQuery(&Query{Needle: ""})
			ctx.Window.Refresh()
		}),
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {}),
		widget.NewToolbarAction(theme.ContentPasteIcon(), func() {
			content := ctx.Window.ClipboardContent()
			title := strings.TrimSuffix(shortText(content, 32), ":")
			note := NewNote(ctx, 0, title)
			note.Set("Body", content+"\n", true)
			currentNotebook := ctx.Window.CurrentWorkingNotebook()

			if currentNotebook == nil {
				dialog.ShowError(errors.New("Please select current working notebook"), ctx.Window.window)
				return
			}

			canWrite, reason := currentNotebook.CanWrite()
			if !canWrite {
				dialog.ShowError(reason, ctx.Window.window)
				return
			}

			if err := currentNotebook.PutData(note); err != nil {
				log.Println(err)
			}
			ctx.Requests <- RequestLoadData
			ctx.Window.Refresh()
		}),
		widget.NewToolbarAction(theme.MediaRecordIcon(), func() {}),
		widget.NewToolbarAction(theme.VisibilityOffIcon(), func() {}),
		widget.NewToolbarAction(theme.DeleteIcon(), func() {
			if ctx.Window.selectedNote == nil {
				return
			}
			warning := fmt.Sprintf("Are you sure you want to delete \"%s\"?",
				ctx.Window.selectedNote.Title)
			dialog.ShowConfirm("", warning, func(yes bool) {
				if yes {
					err := ctx.Window.selectedNote.Source.DeleteData(
						ctx.Window.selectedNote,
					)
					if err != nil {
						dialog.ShowError(err, ctx.Window.window)
					}
					ctx.Window.selectedNote = nil
					ctx.Window.selectedListID = -1
					ctx.Requests <- RequestLoadData
				}
			}, ctx.Window.window)
		}),
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			ctx.Requests <- RequestLoadData
		}),
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			// General tab
			generalTab := container.NewVBox(
				container.New(layout.NewFormLayout(),
					widget.NewLabel("Username:"), widget.NewEntry(),
					widget.NewLabel("Enable Sync:"), widget.NewCheck("", nil),
				),
			)

			// View tab
			viewTab := container.NewVBox(
				container.New(layout.NewFormLayout(),
					widget.NewLabel("Theme:"), widget.NewSelect([]string{"Light", "Dark", "System"}, func(string) {}),
					widget.NewLabel("Font Size:"), widget.NewEntry(),
				),
			)

			// Search tab
			searchTab := container.NewVBox(
				container.New(layout.NewFormLayout(),
					widget.NewLabel("Default Engine:"), widget.NewSelect([]string{"Google", "DuckDuckGo", "Bing"}, func(string) {}),
					widget.NewLabel("Show Suggestions:"), widget.NewCheck("", nil),
				),
			)

			// Notebooks tab
			notebooksTab := container.NewVBox(
				container.New(layout.NewFormLayout(),
					widget.NewLabel("Default Notebook:"), widget.NewEntry(),
					widget.NewLabel("Auto-Save Notes:"), widget.NewCheck("", nil),
				),
			)

			tabs := container.NewAppTabs(
				container.NewTabItem("General", generalTab),
				container.NewTabItem("View", viewTab),
				container.NewTabItem("Search", searchTab),
				container.NewTabItem("Notebooks", notebooksTab),
			)

			// Optional: open as modal-like dialog
			dialog.ShowCustom("Preferences", "Close", tabs, ctx.Window.window)
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

			dialog.ShowCustom("About", "Close", content, ctx.Window.window)

		}),
	)
}
