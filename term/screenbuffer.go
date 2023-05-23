package term

import (
	"log"
	"unicode/utf8"
)

type Screen struct {
	data []rune
	Rows int
	Cols int
}

func NewScreen(r, c int) *Screen {
	scr := &Screen{}
	scr.Rows = r
	scr.Cols = c
	// make([]rune, C*R)
	scr.data = make([]rune, c*r)
	scr.Blank()
	log.Println("created ScreenBuf size", len(scr.data))
	return scr
}

func (scr *Screen) Blank() {
	for i, _ := range scr.data {
		scr.data[i] = ' '
	}
}

func (scr *Screen) Set(c, r int, ch rune) {
	// board[c*C + r] = "abc" // like board[i][j] = "abc"
	//scr.data[(c*scr.Cols)+r] = ch
	scr.data[(r*scr.Rows)+c] = ch
}

func (scr *Screen) Get(c, r int) rune {
	//return scr.data[(c*scr.Cols)+r]
	return scr.data[(r*scr.Rows)+c]
}

func (scr *Screen) GetBytes() []byte {
	buf := make([]byte, len(scr.data)*utf8.UTFMax)

	count := 0
	for _, r := range scr.data {
		count += utf8.EncodeRune(buf[count:], r)
	}
	buf = buf[:count]

	return buf
}
