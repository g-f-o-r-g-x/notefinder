package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
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

	log.Println(appName, appVersion)
	ctx := NewContext()

	toInterp := make(chan *Note, 1)
	worker := NewWorker(ctx, toInterp)
	go worker.Run()

	indexer := &Indexer{context: ctx}
	go indexer.Run()

	/* Initialize within goroutine to lock to thread */
	go func() {
		ctx.interpreter = NewInterpreter()
		defer ctx.interpreter.Destroy()
		ctx.interpreter.Run(toInterp)
	}()
	ctx.Run()
}
