package buffer

import (
	"io/ioutil"
	"log"
	"os"
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
	t.Mods = append(t.Mods, NewPiece(&t.Content, 0, len(c)))
	return t
}

func NewPiece(s *string, pt, r int) *Piece {
	return &Piece{Source: s, Start: pt, Run: r}
}

func (t *Table) size() int {
	i := 0
	for _, n := range t.Mods {
		i += n.Run
	}
	return i
}

func (t *Table) allContents() string {
	//return t.contents(0, t.size())
	// need golang's stringbuilder
	s := ""

	for i := 0; i < len(t.Mods); i++ {
		p := t.Mods[i]
		src := p.Source
		s += string(*src)[p.Start : p.Start+p.Run-1]
	}
	return s
}

func (t *Table) dump() {
	l := log.New(os.Stderr, "", 0)
	l.Println("Table Dump")

	l.Println("Content", t.Content)
	l.Println("Add    ", t.Add)
	for i := 0; i < len(t.Mods); i++ {
		p := t.Mods[i]
		p.dump(l)
	}
}

// need to add observance of start, end
func (t *Table) contents(start, end int) string {
	// need golang's stringbuilder
	s := ""
	si, sp, ss := t.pieceAt(start)
	ei, ep, ee := t.pieceAt(end)

	startFrag := sp.head(ss)
	endFrag := ep.tail(ee)
	s += startFrag
	for i := si + 1; i < ei; i++ {
		p := t.Mods[i]
		src := p.Source
		s += string(*src)[p.Start : p.Start+p.Run]
	}
	s += endFrag
	return s
}

func (p *Piece) size() int {
	return p.Run
}
func (p *Piece) head(idx int) string {
	s := string(*p.Source)
	return s[:idx]
}
func (p *Piece) tail(idx int) string {
	s := string(*p.Source)
	return s[idx:]
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
	e, p, i := t.pieceAt(pt)
	// Appending characters to the "add file" buffer, and
	np := NewPiece(&t.Add, len(t.Add), len(s))
	t.Add += s
	// Updating the entry in piece table (breaking an entry into two or three)
	if i == 0 {
		// insert at p
		t.insertPieceAt(e, np)
		return nil
	}
	if i == p.Run {
		// insert np at p+1
		t.insertPieceAt(e+1, np)
		return nil
	}
	// else split the piece and make { left, np, right }
	left, right := p.splitAt(i)
	t.insertPieceAt(e, left)
	t.insertPieceAt(e+1, right)
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

func (p *Piece) dump(f *log.Logger) {
	f.Println(p.Source, p.Start, p.Run)
}

// splits a piece into two
func (p *Piece) splitAt(idx int) (left, right *Piece) {
	if idx == 0 || idx == p.Run {
		return nil, nil
	}
	left = NewPiece(p.Source, p.Start, idx)
	right = NewPiece(p.Source, idx+1, p.Run-idx)
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
