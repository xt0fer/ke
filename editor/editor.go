package editor

import (
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

	editor.Term = term.NewTerm(term.Pty)
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

func (e *Editor) UpdateTerminal() {
	e.Term.Write([]byte(term.CUP(0, 0)))
	e.Term.Write([]byte(term.ED(term.EraseToEnd)))
	s := e.CurrentScreen()
	e.Term.Write([]byte(s))
}

func (e *Editor) CurrentScreen() string {
	return e.RootBuffer.T.AllContents()
}

func (editor *Editor) DisplayContents(s string) []byte {
	msg := []byte(term.CUP(0, 0))
	msg = append(msg, []byte(term.ED(term.EraseToEnd))...)
	//s := editor.RootBuffer.T.AllContents()
	msg = append(msg, []byte(s)...)
	//log.Println("sizeof display is ", len(msg))
	return msg
}
