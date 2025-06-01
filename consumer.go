package main

import (
	"fmt"
	"sync"
)

type Consumer struct {
	context *Context
	mx      sync.Mutex
}

func NewConsumer(ctx *Context) *Consumer {
	return &Consumer{context: ctx}
}

func (c *Consumer) Run() {
	for {
		note, ok := <-c.context.bus
		if !ok {
			return
		}

		for k, hits := range note.Words() {
			fmt.Printf("\"%s\": %d\n", k, hits)
		}
		fmt.Println("---------------------------------")
	}
}
