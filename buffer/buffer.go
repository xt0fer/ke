package buffer

import (
	"io/ioutil"
	"log"
)

type Table struct {
	Content string
	Add     string
	Mods    []*Piece
}

type Piece struct {
	Source *string
	Start  int
	Run    int
}

func NewTable(c string) *Table {
	t := &Table{Content: c, Add: "", Mods: []*Piece{}}
	t.Mods = append(t.Mods, NewPiece(t.Content, 0, len(c)))
	return t
}

func NewPiece(s string, pt, r int) *Piece {
	return &Piece{Source: &s, Start: pt, Run: r}
}

func (t Table) size() int {
	i := 0
	for _, n := range t.Mods {
		i += n.Run
	}
	return i
}

func (p Piece) size() int {
	return p.Run
}

func (t Table) index(idx int) byte {
	p, i := t.pieceAt(idx)
	src := (p.Source)
	return string(*src)[i]
}

func (t Table) pieceAt(idx int) (*Piece, int) {
	i := idx
	for _, p := range t.Mods {
		if i < p.Run {
			return p, i
		} else {
			i = i - p.Run
		}
	}
	return nil, 0
}

func (t Table) add(s string, pt int) error {
	return nil
}

func (t Table) deleteRune(pt int) {

}

func (p Piece) deleteAt(idx int) {
	if idx == 0 {
		p.Start += 1
	}
	if idx == p.Run {
		p.Run -= 1
	}
	// else split into two pieces
}

func (p Piece) splitAt(idx int) (left, right *Piece) {
	if idx == 0 || idx == p.Run {
		return nil, nil
	}
	left = NewPiece(*p.Source, p.Start, idx)
	right = NewPiece(*p.Source, idx+1, p.Run-idx)
	return
}

// load a text file from a filename string
func LoadFile(filename string) *Table {
	content, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Fatal(err)
	}

	return NewTable(string(content))
}
