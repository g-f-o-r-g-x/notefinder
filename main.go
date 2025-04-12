package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
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

	ch := make(chan int, 1)

	interpreter := NewInterpreter()
	defer interpreter.Destroy()
	go func() {
		for i := range 1000 {
			ch <- i
			time.Sleep(1)
		}
		close(ch)
	}()
	interpreter.Run(ch)
	ctx.Run()
}
