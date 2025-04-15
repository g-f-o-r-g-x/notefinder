package main

/*
#include <EXTERN.h>
#include "perl.h"

void
call_print_note(pTHX_ SV* note_ref)
{
    dSP;                      // initialize stack pointer
    ENTER;
    SAVETMPS;

    PUSHMARK(SP);
    XPUSHs(note_ref);         // push argument
    PUTBACK;

//    call_pv("print_note", G_VOID); // call Perl function

    FREETMPS;
    LEAVE;
}
*/
import "C"

import (
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
		C.CString(""),
		C.CString("-e"),
		C.CString(`sub print_note {
			my $note = shift;
			for my $k (sort keys %$note) {
				print "$k => $note->{$k}\n";
			}
		}`),
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

func (i *Interpreter) Run(input <-chan *Note, toIndex chan<- *Note) {
	for data := range input {
		C.call_print_note(i.perl, data.ToHV())
		time.Sleep(1)

		go func() {
			toIndex <- data
		}()
	}
	close(toIndex)
}
