package editor

import (
	"log"

	"github.com/kristofer/ke/buffer"
	"github.com/kristofer/ke/term"
)

type Editor struct {
	Term      *term.Term
	InputChan chan term.Event
	// CurrentBuffer *Buffer /* current buffer */
	RootBuffer *buffer.Buffer /* head of list of buffers */
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
	e.RootBuffer = buffer.NewBuffer("")
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
	if event.Type == term.EventKey {
		// ctrl-X, ctrl-P does a debugging dump of the PieceTable
		if e.CtrlXFlag && term.Key(event.Ch) == term.KeyCtrlP {
			e.RootBuffer.T.Dump()
			return true
		}
		if e.CtrlXFlag && term.Key(event.Ch) == term.KeyCtrlQ {
			return false
		}
		if term.Key(event.Ch) == term.KeyCtrlX {
			e.CtrlXFlag = true
		} else {
			e.CtrlXFlag = false
		}
		if term.Key(event.Ch) == term.KeyEsc {
			e.EscapeFlag = true
		} else {
			e.EscapeFlag = false
		}
		// else a 'normal' rune
		e.RootBuffer.AddRune(event.Ch)
	}
	return true
}

func (e *Editor) UpdateDisplay() {
	e.Term.Write([]byte(term.CUP(0, 0)))
	e.Term.Write([]byte(term.ED(term.EraseToEnd)))
	s := e.RootBuffer.T.AllContents()
	e.Term.Write([]byte(s))
}
