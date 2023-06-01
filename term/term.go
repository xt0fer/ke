package term

import (
	"bufio"
	"log"
	"os"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/gorilla/websocket"
)

type (
	TermType int
)

const (
	Web TermType = iota
	Pty
)

type Term struct {
	Kind   TermType
	Input  *bufio.Reader
	Output *bufio.Writer
	Origin *syscall.Termios
	// sigwinch       = make(chan os.Signal, 1)
	// sigio          = make(chan os.Signal, 1)
	Quit     chan int
	Contents strings.Builder
	ScrBuf   *Screen
	Conn     *websocket.Conn
	CurCol   int
	CurRow   int
}

// this `term` imeplmentation only really does teh Web type.
// the Pty type of term has not been finished or tested.
func NewTerm(kind TermType) *Term {
	t := &Term{}
	t.Kind = kind

	if kind == Pty {
		t.Input = bufio.NewReader(os.Stdin)
		t.Output = bufio.NewWriter(os.Stdout)
		stdin := os.Stdin.Fd()
		termios := GetTermios(stdin)

		t.Origin = termios

		SetRaw(termios)
		SetTermios(stdin, termios)
		t.Output.Write([]byte(CSIStart()))

	}

	if kind == Web {
		t.ScrBuf = NewScreen(80, 24) // termsize cols, rows
	}

	return t
}

func (t *Term) IsPty() bool { return t.Kind == Pty }
func (t *Term) IsWeb() bool { return t.Kind == Web }

func (t *Term) Cleanup() {
	if t.IsPty() {
		t.Output.Write([]byte(CSIStart()))
		SetTermios(os.Stdin.Fd(), t.Origin)
	}
}

func (t *Term) PollEvent() Event {
	ru, _, err := t.Input.ReadRune()
	//log.Println("Event recv", ru)
	if err != nil {
		panic(err)
	}
	e := Event{}
	e.Type = EventKey
	e.Ch = ru
	return e
}

func (t *Term) EventFromByte(b byte) Event {
	e := Event{}
	e.Type = EventKey
	e.Ch = rune(b)
	return e
}
func (t *Term) EventFromKey(key []byte) Event {
	log.Println("EventFromKey", len(key), key)
	//ru, _, err := t.Input.ReadRune()
	ru, n := utf8.DecodeRune(key)
	if ru == utf8.RuneError && n == 1 {
		log.Println("error on recv", ru)
	}
	e := Event{}
	if len(key) > 1 {
		e.Type = EventKey
		e.Key = StringToKey(string(key))
		e.Ch = 0
		return e
	}
	if (Key(ru) >= KeyCtrlTilde && Key(ru) <= KeySpace) ||
		Key(ru) >= KeyHome && Key(ru) <= KeyArrowRight {
		e.Type = EventKey
		e.Key = Key(ru)
		e.Ch = 0
		//log.Println(e.String())
		return e

	} else {
		e.Type = EventKey
		e.Ch = ru
		//log.Println(e.String())
	}
	return e
}

func (t *Term) Write(b []byte) {
	if t.Kind == Pty {
		t.Output.Write(b)
		t.Output.Flush()
	}
	if t.Kind == Web {
		msgType := 1
		msg := b
		if err := t.Conn.WriteMessage(msgType, msg); err != nil {
			log.Println("unable to write message to frontend")
			return
		}
	}
}

func (t *Term) Blank() {
	t.ScrBuf.Blank()
}

func (t *Term) Flush() {

	if t.IsWeb() {
		//log.Printf("\nOnFlush***\n%s***\n", t.ScrBuf.String())
		msgType := 1
		msg := t.ScrBuf.GetBytes()
		if err := t.Conn.WriteMessage(msgType, msg); err != nil {
			log.Println("unable to write message to frontend")
			return
		}
	}

}

func (t *Term) Clear() {
	t.Blank()
}

func (t *Term) Size() (int, int) {
	return 80, 24 // termsize cols, rows;
}

func (t *Term) SetCell(c, r int, ch rune, fg, bg Attribute) {
	// switch zero-based to one-based?
	t.ScrBuf.Set(c, r, ch)
}
func (t *Term) SetCursor(c int, r int) {
	//log.Println("term.SetCursor", c, r)
	// switch zero-based to one-based?
	t.CurCol = c
	t.CurRow = r
	if t.IsWeb() {
		// t.Write([]byte(CUP(t.CurCol, t.CurRow)))
		t.Write([]byte(CUP(t.CurCol, t.CurRow)))
	}
}
