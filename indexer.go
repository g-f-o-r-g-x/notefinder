package main

import (
	"log"
	"time"
)

type Indexer struct {
	context *Context
}

func (i *Indexer) Run() {
	for {
		log.Println("Indexer is running...")
		time.Sleep(5 * time.Second)
	}
}
