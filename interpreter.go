package main

/*
#include <EXTERN.h>
#include <perl.h>
*/
import "C"

type Interpreter struct {
}

func NewInterpreter() *Interpreter {
	return &Interpreter{}
}

func (i *Interpreter) Run() {
	perl := C.perl_alloc()
	C.perl_construct(perl)
	defer func() {
		C.perl_destruct(perl)
		C.perl_free(perl)
	}()
}
