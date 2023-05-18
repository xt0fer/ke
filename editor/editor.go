package editor

import (
	"github.com/kristofer/ke/term"
)

type Editor struct {
	Term      *term.Term
	InputChan chan term.Event
	// CurrentBuffer *Buffer /* current buffer */
	// RootBuffer    *Buffer /* head of list of buffers */
	// CurrentWindow *Window
	// RootWindow    *Window
	// status vars
	Done bool /* Quit flag. */
	// Msgflag       bool   /* True if msgline should be displayed. */
	// PasteBuffer   string /* Allocated scrap buffer. */
	// Msgline       string /* Message line input/output buffer. */
	// Searchtext    string
	// Replace       string
	//Keymap        []keymapt
	Lines int
	Cols  int
	// FGColor       int
	// BGColor       int
	EscapeFlag    bool
	CtrlXFlag     bool
	MiniBufActive bool
}

func NewEditor() *Editor {
	e := &Editor{}
	e.InputChan = make(chan term.Event, 20)
	return e
}

func (editor *Editor) ForkInputHandler() {

	editor.Term = term.NewTerm()
	defer editor.Term.Cleanup()

	for {
		editor.InputChan <- editor.Term.PollEvent()
	}
}

func (e *Editor) HandleEvent(event term.Event) bool {
	if e.CtrlXFlag && term.Key(event.Ch) == term.KeyCtrlQ {
		return false
	}
	if term.Key(event.Ch) == term.KeyCtrlX {
		e.CtrlXFlag = true
	} else {
		e.CtrlXFlag = false
	}
	return true
}

func (e *Editor) UpdateDisplay() {}
