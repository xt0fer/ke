package term

import (
	"bufio"
	"fmt"
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

// ANSI CSI term codes
// https://en.wikipedia.org/wiki/ANSI_escape_code#CSI_(Control_Sequence_Introducer)_sequences

// Upper Left of screen is 1, 1

const CSI = "\x1b\x5b"

// CSI > 1 u    - start mode
// CSI < u      - end mode

func CSIStart() string {
	return (fmt.Sprintf("%s>1u", CSI))
}
func CSIEnd() string {
	return (fmt.Sprintf("%s<u", CSI))
}

// SHOWCUR - dhow cursor
// func CURBLK() string {
// 	return (fmt.Sprintf("%s0 q", CSI))
// }
// func CURSHOW() string {
// 	return (fmt.Sprintf("%s?25h", CSI))
// }
// func CURHIDE() string {
// 	return (fmt.Sprintf("%s?25l", CSI))
// }

// func DECSET(n int) string {
// 	return (fmt.Sprintf("%s?%dh", CSI, n))
// }
// func DECRESET(n int) string {
// 	return (fmt.Sprintf("%s?%dl", CSI, n))
// }

// CUU - Cursor Up
func CUU(n int) string {
	return (fmt.Sprintf("%s%dA", CSI, n))
}

// CUD - Cursor Down
func CUD(n int) string {
	return (fmt.Sprintf("%s%dB", CSI, n))
}

// CUF - Cursor Forward
func CUF(n int) string {
	return (fmt.Sprintf("%s%dC", CSI, n))
}

// CUB - Cursor Back
func CUB(n int) string {
	return (fmt.Sprintf("%s%dD", CSI, n))
}

// CNL - Cursor Next Line
func CNL(n int) string {
	return (fmt.Sprintf("%s%dE", CSI, n))
}

// CPL - Cursor Previous Line
func CPL(n int) string {
	return (fmt.Sprintf("%s%dF", CSI, n))
}

// CHA - Cursor Horizontal Absolute
func CHA(n int) string {
	return (fmt.Sprintf("%s%dG", CSI, n))
}

// CUP - Cursor Position
func CUP(c, r int) string {
	return (fmt.Sprintf("%s%d;%dH", CSI, r, c))
}

type EraseType int

const (
	EraseToEnd   EraseType = 0
	EraseToBegin EraseType = 1
	EraseAll     EraseType = 2
)

// ED - Erase in Display
// If n is 0, clear from cursor to end of screen.
// If n is 1, clear from cursor to beginning of the screen.
// If n is 2, clear entire screen (and moves cursor to upper left on DOS ANSI.SYS).
func ED(n EraseType) string {
	return (fmt.Sprintf("%s%dJ", CSI, n))
}

// EL - Erase in Line
// If n is 0 (or missing), clear from cursor to the end of the line.
// If n is 1, clear from cursor to beginning of the line.
// If n is 2, clear entire line. Cursor position does not change.
func EL(n EraseType) string {
	return (fmt.Sprintf("%s%dK", CSI, n))
}

// SU - Scroll Up
func SU(n int) string {
	return (fmt.Sprintf("%s%dS", CSI, n))
}

// SD - Scroll Down
func SD(n int) string {
	return (fmt.Sprintf("%s%dT", CSI, n))
}

// HVP - Horizontal Vertical Position
func HVP(m, n int) string {
	return (fmt.Sprintf("%s%d;%df", CSI, n, m))
}

type SGRType int

const (
	SGR_Off          SGRType = 0  // All attributes off
	SGR_Bold         SGRType = 1  // Bold
	SGR_Underline    SGRType = 4  // Underline
	SGR_Blinking     SGRType = 5  // Blinking
	SGR_Negative     SGRType = 7  // Negative image
	SGR_Invisible    SGRType = 8  // Invisible image
	SGR_BoldOff      SGRType = 22 // Bold off
	SGR_UnderlineOff SGRType = 24 // Underline off
	SGR_BlinkingOff  SGRType = 25 // Blinking off
	SGR_NegativeOff  SGRType = 27 // Negative image off
	SGR_InvisibleOff SGRType = 28 // Invisible image off
)

// SGR - Select Graphic Rendition, other data may follow
func SGR(n ...SGRType) string {
	if len(n) < 1 {
		return ""
	}
	s := CSI
	s += fmt.Sprintf("%d", n[0])
	for i := 1; i < len(n); i++ {
		s += fmt.Sprintf(";%d", n[i])
	}
	s += "m"
	return s
}

const (
	SBox_Horiz = 0x2501 //single line
	SBox_Vert  = 0x2503
	SBox_UL    = 0x250F
	SBox_UR    = 0x2513
	SBox_LL    = 0x2517
	SBox_LR    = 0x251B

	DBox_Horiz = 0x2550 //double line
	DBox_Vert  = 0x2551
	DBox_UL    = 0x2554
	DBox_UR    = 0x2557
	DBox_LL    = 0x255A
	DBox_LR    = 0x255D

	BBox_Horiz = 0x2509 //broken line
	BBox_Vert  = 0x250B
	BBox_UL    = 0x250F
	BBox_UR    = 0x2513
	BBox_LL    = 0x2517
	BBox_LR    = 0x251B

	HBox_Horiz = 0x2501 //horiz only
	HBox_Vert  = 0x20
	HBox_UL    = 0x20
	HBox_UR    = 0x20
	HBox_LL    = 0x20
	HBox_LR    = 0x20
)
