package main

import (
	"log"
	"time"
)

type Indexer struct {
	context *Context
}

func NewIndexer(ctx *Context) *Indexer {
	return &Indexer{context: ctx}
}

func (i *Indexer) Run() {
	for {
		log.Println("Indexer is running...")
		time.Sleep(5 * time.Second)
		return
	}
}
