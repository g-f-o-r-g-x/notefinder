package main

import (
	"context"
	"log"
	"os/user"
	"path/filepath"
	"strings"

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
	Notebooks     map[string]*Notebook
	Data          *Store
	CommonStorage *CommonStorage

	Application fyne.App
	Window      *Window
	Interpreter *Interpreter

	Bus      chan *Note
	Requests chan Request
}

func NewContext(Interpreter *Interpreter) *Context {
	a := app.NewWithID("org.notefinder.app")
	ctx := &Context{
		base: base{
			context: context.Background(),
		},
		Interpreter: Interpreter,
		Application: a,
		Bus:         make(chan *Note, 1),
		Requests:    make(chan Request, 1),
	}

	ctx.CommonStorage = NewCommonStorage(ctx)
	ctx.Data = NewStore(ctx)
	ctx.Notebooks = readConfig(ctx)
	ctx.Window = NewWindow(ctx)
	return ctx
}

func (ctx *Context) Run() int {
	log.Println(appName, appVersion, "started")
	log.Println(strings.Repeat("-", len(appName)+12))

	ctx.Window.Show()
	close(ctx.Requests)
	return 0
}

func (ctx *Context) Log(l ...any) {
	log.Println(l...)
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
	case "google":
		return NewGoogleImplementation(ctx, config)
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
