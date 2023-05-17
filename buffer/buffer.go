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
// Mods ends up being a series of 'runs' which contain the current
// value of the document

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

func (t *Table) size() int {
	i := 0
	for _, n := range t.Mods {
		i += n.Run
	}
	return i
}

func (t *Table) allContents() string {
    return t.contents(0,t.size())
}
func (t *Table) contents(start, end int) string {
    // need golang's stringbuilder
    s := ""
	for _, p := range t.Mods {
        src := (p.Source)
		s += string(*src)[p.Start : p.Start+p.Run]
	}
	return s
}

func (p *Piece) size() int {
	return p.Run
}

func (t *Table) index(idx int) byte {
	_, p, i := t.pieceAt(idx)
	src := (p.Source)
	return string(*src)[i]
}

func (t *Table) appendPiece(p *Piece) {
	t.insertPieceAt(len(t.Mods), p)
}

func (t *Table) insertPieceAt(index int, p *Piece) {
	if index >= len(t.Mods) {
		t.Mods = append(t.Mods[:], p)
		return
	}
	t.Mods = append(t.Mods[:index+1], t.Mods[index:]...)
	t.Mods[index] = p
	return
}

// func insert(a []int, index int, value int) []int {
//     a = append(a[:index+1], a[index:]...) // Step 1+2
//     a[index] = value                      // Step 3
//     return a
// }

func (t *Table) pieceAt(idx int) (int, *Piece, int) {
	i := idx
	for j, p := range t.Mods {
		if i < p.Run {
			return j, p, i
		} else {
			i = i - p.Run
		}
	}
	return 0, nil, 0
}

func (t *Table) add(s string, pt int) error {
	return nil
}

func (t *Table) deleteRune(idx int) {
    which, p, i := t.pieceAt(idx)
    if i == 0 || i == p.Run {
        p.trimRune(i)
    }
    // else split into two pieces
    left, right := p.splitAt(i)
    left.Run -= 1 // trim last rune of left (orig idx above)
    t.insertPieceAt(which, right)
    t.insertPieceAt(which, left)
}

func (p *Piece) trimRune(idx int) {
	if idx == 0 {
		p.Start += 1
	}
	if idx == p.Run {
		p.Run -= 1
	}
}
// splits a piece into two
func (p *Piece) splitAt(idx int) (left, right *Piece) {
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
