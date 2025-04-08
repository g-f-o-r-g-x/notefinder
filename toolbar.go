package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"

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
			ctx.MainWindow.SetQuery(&Query{Needle: ""})
			ctx.MainWindow.Refresh()
		}),
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {}),
		widget.NewToolbarAction(theme.MediaRecordIcon(), func() {}),
		widget.NewToolbarAction(theme.VisibilityOffIcon(), func() {}),
		widget.NewToolbarAction(theme.DeleteIcon(), func() {
			if ctx.MainWindow.selectedNote == nil {
				return
			}
			warning := fmt.Sprintf("Are you sure you want to delete \"%s\"?",
				ctx.MainWindow.selectedNote.Title)
			dialog.ShowConfirm("", warning, func(yes bool) {
				if yes {
					err := ctx.MainWindow.selectedNote.Source.DeleteData(
						ctx.MainWindow.selectedNote,
					)
					if err != nil {
						dialog.ShowError(err, ctx.MainWindow.window)
					}
					ctx.MainWindow.selectedNote = nil
					ctx.MainWindow.selectedListID = -1
					ctx.Requests <- RequestLoadData
				}
			}, ctx.MainWindow.window)
		}),
		widget.NewToolbarAction(theme.ContentPasteIcon(), func() {
			content := ctx.MainWindow.ClipboardContent()
			note := NewNote(0, shortText(content, 32), content+"\n")
			currentNotebook := ctx.MainWindow.CurrentWorkingNotebook()

			if currentNotebook == nil {
				dialog.ShowError(errors.New("Please select current working notebook"), ctx.MainWindow.window)
				return
			}

			canWrite, reason := currentNotebook.CanWrite()
			if !canWrite {
				dialog.ShowError(reason, ctx.MainWindow.window)
				return
			}

			if err := currentNotebook.PutData(note); err != nil {
				log.Println(err)
			}
			ctx.Requests <- RequestLoadData
			ctx.MainWindow.Refresh()
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
