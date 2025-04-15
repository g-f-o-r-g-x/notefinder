package main

import (
	"fmt"
	// "time"
)

type Indexer struct {
	context *Context
}

func NewIndexer(ctx *Context) *Indexer {
	return &Indexer{context: ctx}
}

func (i *Indexer) Run(toIndex <-chan *Note) {
	for {
		note := <-toIndex
		for k, hits := range note.Words() {
			fmt.Printf("\"%s\": %d\n", k, hits)
		}
		fmt.Println("---------------------------------")
		//time.Sleep(5 * time.Second)
	}
}
