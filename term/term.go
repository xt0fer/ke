package term

import (
	"bufio"
	"os"
	"syscall"
)

type Term struct {
	Input  *bufio.Reader
	Output *bufio.Writer
	Origin *syscall.Termios
	// sigwinch       = make(chan os.Signal, 1)
	// sigio          = make(chan os.Signal, 1)
	Quit chan int
}

func NewTerm() *Term {
	t := &Term{}
	t.Input = bufio.NewReader(os.Stdin)
	t.Output = bufio.NewWriter(os.Stdout)
	stdin := os.Stdin.Fd()
	termios := GetTermios(stdin)

	t.Origin = termios

	SetRaw(termios)
	SetTermios(stdin, termios)

	t.Output.Write([]byte(CSIStart()))
	return t
}

func (t *Term) Cleanup() {
	t.Output.Write([]byte(CSIStart()))
	SetTermios(os.Stdin.Fd(), t.Origin)
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

func (t *Term) Write(b []byte) {
	t.Output.Write(b)
	t.Output.Flush()
}
func Flush() {

}
