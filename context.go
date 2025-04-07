package main

import (
	"context"
	"os/user"
	"path/filepath"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"gopkg.in/ini.v1"
)

type base struct {
	context context.Context
}

type Request int

const (
	RequestLoadData Request = iota
	RequestStop
)

type Context struct {
	base
	Config      map[string]string
	Notebooks   map[string]*Notebook
	Data        *Store
	Application fyne.App
	MainWindow  *Window

	Requests chan Request
}

func NewContext() *Context {
	a := app.NewWithID("org.notefinder.app")
	ctx := &Context{
		base: base{
			context: context.Background(),
		},
		Config:      readConfig(),
		Data:        NewStore(),
		Application: a,
		Requests:    make(chan Request, 1),
	}

	ctx.MainWindow = NewWindow(ctx)
	return ctx
}

func (ctx *Context) Run() {
	ctx.MainWindow.Show()
}

func getAbsolutePath() string {
	user, _ := user.Current()

	return filepath.Join(user.HomeDir, configPath)
}

func readConfig() map[string]string {
	cfg, err := ini.Load(getAbsolutePath())

	if err != nil {
		panic(err)
	}

	return map[string]string{"path": cfg.Section("default").Key("path").String()}
}
