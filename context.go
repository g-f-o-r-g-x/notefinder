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
	Notebooks   map[string]*Notebook
	Data        *Store
	Application fyne.App
	MainWindow  *Window
	interpreter *Interpreter
	Requests    chan Request
}

func NewContext() *Context {
	a := app.NewWithID("org.notefinder.app")
	ctx := &Context{
		base: base{
			context: context.Background(),
		},
		Application: a,
		Requests:    make(chan Request, 1),
	}

	ctx.Data = NewStore(ctx)
	ctx.Notebooks = readConfig(ctx)
	ctx.MainWindow = NewWindow(ctx)
	return ctx
}

func (ctx *Context) Run() {
	ctx.MainWindow.Show()
	close(ctx.Requests)
}

func getAbsolutePath() string {
	user, _ := user.Current()

	return filepath.Join(user.HomeDir, configPath)
}

func implByName(ctx *Context, name string, config map[string]string) Implementation {
	switch name {
	case "file":
		return NewFileImplementation(ctx, config)
	case "mozilla":
		return NewMozillaImplementation(ctx, config)
	default:
		return nil
	}
}

func readConfig(ctx *Context) map[string]*Notebook {
	cfg, err := ini.Load(getAbsolutePath())

	if err != nil {
		panic(err)
	}

	sections := cfg.Sections()
	ret := make(map[string]*Notebook, len(sections))

	for _, section := range sections {
		config := make(map[string]string)
		name := section.Name()
		if name == "DEFAULT" {
			continue
		}
		for _, key := range section.KeyStrings() {
			config[key] = section.Key(key).String()
		}

		implName, haveImplName := config["impl"]
		_, havePath := config["path"]

		if !haveImplName && havePath {
			implName = "file"
		}
		impl := implByName(ctx, implName, config)
		_ = impl

		ret[name] = NewNotebook(name, impl, config, NotebookConfigured)

	}

	return ret
}
