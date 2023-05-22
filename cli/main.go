package main

import (
	"github.com/kristofer/ke/editor"
	"github.com/kristofer/ke/term"
)

func main() {

	editor := editor.NewEditor()

	go web.echoserver()

	go editor.ForkInputHandler()
	// Instead of using for {
	// 	select {
	// 	case ev := <-e.EventChan:

	for event := range editor.InputChan {
		ok := editor.HandleEvent(event)
		if !ok {
			return //exit editor
		}
		editor.UpdateDisplay()
		term.Flush()
	}

}
