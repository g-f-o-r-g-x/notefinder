//go:build darwin

package interpreter

type Interpreter struct {
}

func NewInterpreter() *Interpreter {
	return &Interpreter{}
}

func (i *Interpreter) Eval(code string) {
}

func (i *Interpreter) Destroy() {
}

func (i *Interpreter) Call(fn string, args ...struct{}) {
}
