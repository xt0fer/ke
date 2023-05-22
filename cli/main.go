package main

import (
	"github.com/kristofer/ke/editor"
	"github.com/kristofer/ke/term"
)

// this is for running in a pty/tty environment
// see web/ package for the JS frontend for ke
func main() {

	editor := editor.NewEditor()

	go editor.ForkInputHandler()

	for event := range editor.InputChan {
		ok := editor.HandleEvent(event)
		if !ok {
			return //exit editor
		}
		editor.UpdateTerminal()
		term.Flush()
	}

}
