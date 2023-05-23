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
	CurCol int
	CurRow int
}

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
		t.ScrBuf = NewScreen(24, 80)
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

// func (t *Term) Truncate() {
// 	t.Contents.Reset()
// }

// func (t *Term) AppendBytes(buf []byte) {
// 	t.Contents.Write(buf)
// }

// func (t *Term) Append(s string) {
// 	t.Contents.Write([]byte(s))
// }

// func (t *Term) GetContents() []byte {
// 	return []byte(t.Contents.String())
// }

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

func (t *Term) EventFromKey(key []byte) Event {
	//ru, _, err := t.Input.ReadRune()
	ru, n := utf8.DecodeRune(key)
	if ru == utf8.RuneError && n == 1 {
		log.Println("error on recv", ru)
	}
	//log.Println("Event recv", ru)
	// if err != nil {
	// 	panic(err)
	// }
	e := Event{}
	e.Type = EventKey
	e.Ch = ru
	return e
}

func (t *Term) Write(b []byte) {
	if t.Kind == Pty {
		t.Output.Write(b)
		t.Output.Flush()
	}
	if t.Kind == Web {

	}
}

func (t *Term) Flush() {

	if t.IsWeb() {
		msgType := 1;
		msg := t.ScrBuf.GetBytes()
		if err := t.Conn.WriteMessage(msgType, msg); err != nil {
			log.Println("unable to write message to frontend")
			return
		}
	}

}
