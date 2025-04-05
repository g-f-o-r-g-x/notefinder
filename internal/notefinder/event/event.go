package notefinder

import (
	_ "fmt"
)

func eventOnLoaded(note *Note) {
	//	fmt.Println(note)
	//	note.context.interpreter.Call("Notefinder::Hook::OnNoteLoaded", note.ToHV())
}
