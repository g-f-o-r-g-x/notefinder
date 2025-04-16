package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"strings"
)

const (
	doProfiling = false
)

func main() {
	if doProfiling {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	log.Println(appName, appVersion, "started")
	log.Println(strings.Repeat("-", len(appName)+12))
	ctx := NewContext()

	toInterp := make(chan *Note, 1)
	toIndex := make(chan *Note, 1)
	worker := NewWorker(ctx, toInterp)
	go worker.Run()

	/* Initialize within goroutine to lock to thread */
	go func() {
		ctx.interpreter = NewInterpreter(ctx)
		defer ctx.interpreter.Destroy()
		ctx.interpreter.Run(toInterp, toIndex)
	}()
	indexer := &Indexer{context: ctx}
	go indexer.Run(toIndex)

	ctx.Run()
}
