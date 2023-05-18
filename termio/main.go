//go:build darwin || linux
// +build darwin linux

package main

import (
	"os"
	"syscall"
	"unsafe"
)

func getTermios(fd uintptr) *syscall.Termios {
	var t syscall.Termios
	_, _, err := syscall.Syscall6(
		syscall.SYS_IOCTL,
		os.Stdin.Fd(),
		syscall.TIOCGETA, //TCGETS,
		uintptr(unsafe.Pointer(&t)),
		0, 0, 0)

	if err != 0 {
		panic("err")
	}

	return &t
}

func setTermios(fd uintptr, term *syscall.Termios) {
	_, _, err := syscall.Syscall6(
		syscall.SYS_IOCTL,
		os.Stdin.Fd(),
		syscall.TIOCSETA, //TCSETS,
		uintptr(unsafe.Pointer(term)),
		0, 0, 0)
	if err != 0 {
		panic("err")
	}
}

func setRaw(term *syscall.Termios) {
	// This attempts to replicate the behaviour documented for cfmakeraw in
	// the termios(3) manpage.
	term.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP | syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	// newState.Oflag &^= syscall.OPOST
	term.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	term.Cflag &^= syscall.CSIZE | syscall.PARENB
	term.Cflag |= syscall.CS8

	term.Cc[syscall.VMIN] = 1
	term.Cc[syscall.VTIME] = 0
}

func main() {
	t := getTermios(os.Stdin.Fd())

	origin := *t
	defer func() {
		setTermios(os.Stdin.Fd(), &origin)
	}()

	setRaw(t)
	setTermios(os.Stdin.Fd(), t)

	for i := 0; i < 3; i++ {
		buf := make([]byte, 1)
		syscall.Read(0, buf)
		println(buf[0])
	}
}
