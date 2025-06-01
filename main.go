package main

import "os"

func main() {
	interpreter := NewInterpreter()
	defer interpreter.Destroy()
	ctx := NewContext(interpreter)

	go NewWorker(ctx).Run()
	go NewConsumer(ctx).Run()
	os.Exit(ctx.Run())
}
