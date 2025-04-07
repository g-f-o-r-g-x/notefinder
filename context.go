package main

import (
	"context"
	"log"
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
	//	Config      map[string]string
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
		//		Config:      readConfig(),
		Data:        NewStore(),
		Application: a,
		Requests:    make(chan Request, 1),
	}

	ctx.Notebooks = readConfig(ctx)
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

func implByName(name string, config map[string]string) Implementation {
	switch name {
	case "file":
		return NewFileImplementation(config)
	case "mozilla":
		return NewMozillaImplementation(config)
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

		log.Println(config)

		implName, haveImplName := config["impl"]
		_, havePath := config["path"]

		if !haveImplName && havePath {
			implName = "file"
		}
		impl := implByName(implName, config)
		_ = impl

		ret[name] = NewNotebook(name, impl, config, NotebookConfigured)

	}

	return ret
	// return map[string]string{"path": cfg.Section("default").Key("path").String()}
}
