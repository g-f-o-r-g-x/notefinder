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

	ch := make(chan *Note, 1)
	go worker(ctx, ch)
	indexer := &Indexer{context: ctx}
	go indexer.Run()

	go func() {
		ctx.interpreter = NewInterpreter()
		defer ctx.interpreter.Destroy()
		ctx.interpreter.Run(ch)
	}()
	ctx.Run()
}
