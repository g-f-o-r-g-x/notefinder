package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	log.Println(appName, appVersion)
	ctx := NewContext()

	go load(ctx)
	ctx.Run()
}
