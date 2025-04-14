package main

import (
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
		time.Sleep(5 * time.Second)
		return
	}
}
