package main

import (
	"log"
)

func main() {
	log.Println(appName, appVersion)
	ctx := NewContext()

	go load(ctx)
	ctx.Run()
}
