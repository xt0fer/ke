package term

import (
	"bufio"
	"os"
	"syscall"
)

type Term struct {
	Input  *bufio.Reader
	Origin *syscall.Termios
}

func NewTerm() *Term {
	t := &Term{}
	t.Input = bufio.NewReader(os.Stdin)
	stdin := os.Stdin.Fd()
	termios := GetTermios(stdin)

	t.Origin = termios

	SetRaw(termios)
	SetTermios(stdin, termios)

	return t
}

func (t *Term) Cleanup() {
	SetTermios(os.Stdin.Fd(), t.Origin)
}

func (t *Term) PollEvent() Event {
	ru, _, err := t.Input.ReadRune()
	if err != nil {
		panic(err)
	}
	e := Event{}
	e.Ch = ru
	return e
}

func Flush() {

}
