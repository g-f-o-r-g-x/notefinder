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

	go worker(ctx)
	indexer := &Indexer{context: ctx}
	go indexer.Run()
	ctx.Run()
}
