package main

/*
#include <EXTERN.h>
#include "perl.h"
*/
import "C"

import (
	"fmt"
	"runtime"
	"time"
	"unsafe"
)

type Interpreter struct {
	perl *C.PerlInterpreter
}

func NewInterpreter() *Interpreter {
	runtime.LockOSThread()

	args := []*C.char{
		C.CString(""), C.CString("-e"), C.CString("0"), nil,
	}
	defer func() {
		for _, arg := range args {
			if arg != nil {
				C.free(unsafe.Pointer(arg))
			}
		}
	}()
	argc := C.int(len(args) - 1)
	argv := (**C.char)(unsafe.Pointer(&args[0]))
	perl := C.perl_alloc()
	C.perl_construct(perl)
	C.perl_parse(perl, nil, argc, argv, nil)
	C.perl_run(perl)

	return &Interpreter{perl: perl}
}

func (i *Interpreter) Eval(code string) {
	cCode := C.CString(code)
	defer C.free(unsafe.Pointer(cCode))
	C.Perl_eval_pv(i.perl, cCode, 1)
}

func (i *Interpreter) Destroy() {
	C.perl_destruct(i.perl)
	C.perl_free(i.perl)
	runtime.UnlockOSThread()
}

func (i *Interpreter) Run(input <-chan int) {
	for num := range input {
		code := fmt.Sprintf("print(\"Go sent: %d\n\");", num)
		i.Eval(code)
		time.Sleep(1)
	}
}
