package ui

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

	"notefinder/internal/notefinder/common"
	"notefinder/internal/notefinder/types"
	"notefinder/internal/notefinder/util"
)

func makeToolbar(win *Window) *widget.Toolbar {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.HomeIcon(), func() {
			win.SetQuery(&types.Query{Needle: ""})
			win.Refresh()
		}),
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			openNote(win, -1)
		}),
		widget.NewToolbarAction(theme.ContentPasteIcon(), func() {
			content := win.ClipboardContent()
			title := strings.TrimSuffix(util.ShortText(content, 32), ":")
			note := types.NewNote(0, title)
			note.Set("Body", content+"\n", true)
			currentNotebook := win.CurrentWorkingNotebook()

			if currentNotebook == nil {
				dialog.ShowError(errors.New("Please select current working notebook"), win)
				return
			}

			canWrite, reason := currentNotebook.CanWrite()
			if !canWrite {
				dialog.ShowError(reason, win)
				return
			}

			if err := currentNotebook.PutData(note); err != nil {
				log.Println(err)
			}
			win.RequestRefresh()
			win.Refresh()
		}),
		widget.NewToolbarAction(theme.MediaRecordIcon(), func() {}),
		widget.NewToolbarAction(theme.VisibilityOffIcon(), func() {}),
		widget.NewToolbarAction(theme.DeleteIcon(), func() {
			if win.selectedNote == nil {
				return
			}
			warning := fmt.Sprintf("Are you sure you want to delete \"%s\"?",
				win.selectedNote.Title)
			dialog.ShowConfirm("", warning, func(yes bool) {
				if yes {
					err := win.selectedNote.Source.DeleteData(
						win.selectedNote,
					)
					if err != nil {
						dialog.ShowError(err, win)
					}
					win.selectedNote = nil
					win.selectedListID = -1
					win.RequestRefresh()
				}
			}, win)
		}),
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			win.RequestRefresh()
		}),
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			generalTab := container.NewVBox(
				container.New(layout.NewFormLayout(),
					widget.NewLabel("Username:"), widget.NewEntry(),
					widget.NewLabel("Enable Sync:"), widget.NewCheck("", nil),
				),
			)

			viewTab := container.NewVBox(
				container.New(layout.NewFormLayout(),
					widget.NewLabel("Theme:"), widget.NewSelect([]string{"Light", "Dark", "System"}, func(string) {}),
					widget.NewLabel("Font Size:"), widget.NewEntry(),
				),
			)

			searchTab := container.NewVBox(
				container.New(layout.NewFormLayout(),
					widget.NewLabel("Default Engine:"), widget.NewSelect([]string{"Google", "DuckDuckGo", "Bing"}, func(string) {}),
					widget.NewLabel("Show Suggestions:"), widget.NewCheck("", nil),
				),
			)

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

			dialog.ShowCustom("Preferences", "Close", tabs, win)
		}),
		widget.NewToolbarAction(theme.InfoIcon(), func() {
			img := canvas.NewImageFromResource(&fyne.StaticResource{
				StaticName:    "notefinder.png",
				StaticContent: logo,
			})
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(128, 128))

			headerText := fmt.Sprintf("%s %.1f", common.AppName, common.AppVersion)
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

			dialog.ShowCustom("About", "Close", content, win)

		}),
	)
}
