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

	call_pv("print_note", G_VOID); // call Perl function
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
	context *Context
	perl *C.PerlInterpreter
}

func NewInterpreter(context *Context) *Interpreter {
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

//	SV **sp = vTHX->Istack_sp;
//	push_scope();
//	Perl_savetmps(aTHX);
//	do {
//        Stack_off_t * mark_stack_entry;                               \
//        if (UNLIKELY((mark_stack_entry = ++PL_markstack_ptr)          \
//                                           == PL_markstack_max)) {      \
//            mark_stack_entry = Perl_markstack_grow(aTHX);                 }     \
//        *mark_stack_entry  = (Stack_off_t)((sp) - PL_stack_base);      \
//
//      } while 0;
//	do { EXTEND(sp,1); PUSHs(s); } while 0;
//      PL_stack_sp = sp;
// Perl_call_pv(aTHX_ a,b);
// if (PL_tmps_ix > PL_tmps_floor) Perl_free_tmps(aTHX);
// Perl_pop_scope(aTHX);

func (i *Interpreter) Call(fn string, args ...*C.SV) {
	i.context.Log("entering", currentFunction())
	stackPtr := i.perl.Istack_sp

//	C.Perl_push_scope(i.perl)
//	C.Perl_savetmps(i.perl)

	i.context.Log("stack pointer mark")
	i.perl.Imarkstack_ptr = (*C.Stack_off_t)(unsafe.Pointer(
		uintptr(unsafe.Pointer(i.perl.Imarkstack_ptr)) + unsafe.Sizeof(*i.perl.Imarkstack_ptr)))
	mark := i.perl.Imarkstack_ptr
	if mark == i.perl.Imarkstack_max {
		i.context.Log("growing markstack")
		mark = C.Perl_markstack_grow(i.perl)
	}

	/* Push args to stack and move stack pointer */
	i.context.Log("pushing args")
	nArgs := len(args)
	stackPtr = C.Perl_stack_grow(i.perl, stackPtr, stackPtr, C.SSize_t(nArgs))
	i.context.Log("growing stack")
	for _, arg := range args {
		stackPtr = (**C.SV)(unsafe.Pointer(
		uintptr(unsafe.Pointer(i.perl.Istack_sp)) + unsafe.Sizeof(*i.perl.Istack_sp)))
		*stackPtr = arg
		i.perl.Istack_sp = stackPtr
	}

	cFn := C.CString(fn)
	defer C.free(unsafe.Pointer(cFn))
	i.context.Log("calling function")
	C.Perl_call_pv(i.perl, cFn, C.G_VOID)
	i.context.Log("freeing tmps")
//	if i.perl.Itmps_ix > i.perl.Itmps_floor {
//		C.Perl_free_tmps(i.perl)
//	}
//	C.Perl_pop_scope(i.perl)
	i.context.Log("returning from", currentFunction())
}

func (i *Interpreter) Run(input <-chan *Note, toIndex chan<- *Note) {
	i.context.Log("entering", currentFunction())
	for data := range input {
		i.Call("print_note", data.ToHV())
		//C.call_print_note(i.perl, data.ToHV())
		time.Sleep(1)

		go func() {
			toIndex <- data
		}()
	}
	close(toIndex)
}
