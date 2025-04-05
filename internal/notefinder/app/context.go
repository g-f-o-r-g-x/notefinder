package app

import (
	"context"
	"log"
	"os/user"
	"path/filepath"
	"strings"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"gopkg.in/ini.v1"

	"notefinder/internal/notefinder/background"
	"notefinder/internal/notefinder/common"
	"notefinder/internal/notefinder/db"
	"notefinder/internal/notefinder/implementation"
	"notefinder/internal/notefinder/interpreter"
	"notefinder/internal/notefinder/types"
	"notefinder/internal/notefinder/ui"
)

type Context struct {
	context.Context

	Application fyne.App
	Window      *ui.Window
	Interpreter *interpreter.Interpreter

	Worker   *background.Worker
	Consumer *background.Consumer

	Data          *Store
	CommonStorage *db.CommonStorage

	Bus      chan *types.Note
	Requests chan common.Request
}

func NewContext(interpreter *interpreter.Interpreter) *Context {
	ctx := &Context{
		Context:     context.Background(),
		Interpreter: interpreter,
		Application: app.NewWithID("org.notefinder.app"),
		Bus:         make(chan *types.Note, 1),
		Requests:    make(chan common.Request, 1),
	}

	ctx.Data = NewStore(ctx)
	ctx.Worker = background.NewWorker(ctx, ctx.Data)
	ctx.Consumer = background.NewConsumer(ctx)
	ctx.CommonStorage = db.NewCommonStorage()
	ctx.Window = ui.NewWindow(ctx, ctx.Data, ctx.Application)
	return ctx
}

func (ctx *Context) ReadBus() (*types.Note, bool) {
	note, ok := <-ctx.Bus
	return note, ok
}

func (ctx *Context) CloseBus() {
	close(ctx.Bus)
}

func (ctx *Context) WriteToBus(note *types.Note) {
	ctx.Bus <- note
}

func (ctx *Context) ReadRequest() common.Request {
	req := <-ctx.Requests

	return req
}

func (ctx *Context) WriteRequest(req common.Request) {
	ctx.Requests <- req
}

func (ctx *Context) GetRequests() chan common.Request {
	return ctx.Requests
}

func (ctx *Context) Refresh() {
	ctx.Window.Refresh()
}

func (ctx *Context) Run() int {
	log.Println(common.AppName, common.AppVersion, "started")
	log.Println(strings.Repeat("-", len(common.AppName)+12))

	go ctx.Worker.Run()
	go ctx.Consumer.Run()

	ctx.Window.Show()
	close(ctx.Requests)
	ctx.Interpreter.Destroy()
	return 0
}

func getAbsolutePath() string {
	user, _ := user.Current()

	return filepath.Join(user.HomeDir, common.ConfigPath)
}

func implByName(ctx *Context, name string, config map[string]string) types.Implementation {
	switch name {
	case "file":
		return implementation.NewFileImplementation(config)
	case "mozilla":
		return implementation.NewMozillaImplementation(config)
	case "google":
		return implementation.NewGoogleImplementation(config)
	default:
		return nil
	}
}

func readConfig(ctx *Context) map[string]*types.Notebook {
	cfg, err := ini.Load(getAbsolutePath())

	if err != nil {
		panic(err)
	}

	sections := cfg.Sections()
	ret := make(map[string]*types.Notebook, len(sections))

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

		ret[name] = types.NewNotebook(name, impl, config, types.NotebookConfigured)

	}

	return ret
}
