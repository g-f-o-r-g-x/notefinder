package main

import fyne "fyne.io/fyne/v2"

const (
	appName    = "Notefinder"
	appVersion = 0.1
	configPath = ".config/notefinder.ini"
)

var appLogo = &fyne.StaticResource{
	StaticName:    "notefinder.png",
	StaticContent: logo,
}
