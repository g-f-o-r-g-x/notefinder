package main

/*
#include <EXTERN.h>
#include "perl.h"
*/
import "C"

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
	"unsafe"
)

type Interpreter struct {
	context *Context
	perl    *C.PerlInterpreter
}

func NewInterpreter(context *Context) *Interpreter {
	runtime.LockOSThread()

	home, err := os.UserHomeDir()
	if err != nil {
		panic("Cannot find user home directory: " + err.Error())
	}
	scriptPath := filepath.Join(home, ".config", "Notefinder.pl")
	cScript := C.CString(scriptPath)

	args := []*C.char{
		C.CString(""),
		cScript,
		nil,
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

func (i *Interpreter) Call(fn string, args ...*C.SV) {
	cFn := C.CString(fn)
	defer C.free(unsafe.Pointer(cFn))
	cv := C.Perl_get_cv(i.perl, cFn, 0)
	if cv == nil {
		log.Println(fmt.Errorf("cannot find Perl subroutine \"%s\"", fn))
		return
	}

	stackPtr := i.perl.Istack_sp // dSP
	C.Perl_push_scope(i.perl)    // ENTER
	C.Perl_savetmps(i.perl)      // SAVETMPS

	i.perl.Imarkstack_ptr = (*C.Stack_off_t)(
		unsafe.Pointer(uintptr(unsafe.Pointer(i.perl.Imarkstack_ptr)) +
			unsafe.Sizeof(*i.perl.Imarkstack_ptr))) // PUSHMARK(SP)
	markStackPtr := i.perl.Imarkstack_ptr
	if markStackPtr == i.perl.Imarkstack_max {
		markStackPtr = C.Perl_markstack_grow(i.perl)
	}
	*markStackPtr = (C.Stack_off_t)(uintptr(unsafe.Pointer(stackPtr))-
		uintptr(unsafe.Pointer(i.perl.Istack_base))) /
		(C.Stack_off_t)(unsafe.Sizeof(*stackPtr))

	stackPtr = C.Perl_stack_grow(i.perl, stackPtr, stackPtr,
		C.SSize_t(len(args))) // EXTEND(SP, n)

	for _, arg := range args { // PUSHs(...)
		stackPtr = (**C.SV)(unsafe.Pointer(uintptr(unsafe.Pointer(stackPtr)) +
			unsafe.Sizeof(*stackPtr)))
		*stackPtr = arg
	}
	i.perl.Istack_sp = stackPtr // PUTBACK

	C.Perl_call_sv(i.perl, (*C.SV)(unsafe.Pointer(cv)), C.G_VOID)

	C.Perl_free_tmps(i.perl) // FREETMPS
	C.Perl_pop_scope(i.perl) // LEAVE
}

func (i *Interpreter) Run(input <-chan *Note, toIndex chan<- *Note) {
	i.context.Log("entering", currentFunction())
	for data := range input {
		i.Call("Notefinder::Hook::OnNoteLoaded", data.ToHV())
		time.Sleep(1)

		go func() {
			toIndex <- data
		}()
	}
	close(toIndex)
}
