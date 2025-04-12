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

	go func() {
		for i := range 1000 {
			ch <- i
			time.Sleep(1 * time.Second)
		}
		close(ch)
	}()
	go func() {
		interpreter := NewInterpreter()
		defer interpreter.Destroy()
		interpreter.Run(ch)
	}()
	ctx.Run()
}
