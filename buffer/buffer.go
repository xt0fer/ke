package buffer

import (
	"io/ioutil"
	"log"
	"os"
)

type PieceSource int

const (
	Content PieceSource = 0
	Add     PieceSource = 1
)

type Table struct {
	Content string
	Add     string
	Mods    []*Piece
}

// Mods ends up being a series of 'runs' which contain the current
// value of the document

type Piece struct {
	Source PieceSource
	Start  int
	Run    int
}

func NewTable(c string) *Table {
	t := &Table{Content: c, Add: "", Mods: []*Piece{}}
	t.Mods = append(t.Mods, NewPiece(Content, 0, len(c)))
	return t
}

func NewPiece(s PieceSource, pt, r int) *Piece {
	return &Piece{Source: s, Start: pt, Run: r}
}

func (t *Table) size() int {
	i := 0
	for _, n := range t.Mods {
		i += n.Run
	}
	return i
}

func (t *Table) source(ps PieceSource) string {
	if ps == Content {
		return t.Content
	} else {
		return t.Add
	}
}

func (t *Table) runForMod(index int) string {
	p := t.Mods[index]
	return t.source(p.Source)[p.Start : p.Start+p.Run]
}

func (t *Table) head(p *Piece, idx int) string {
	return t.source(p.Source)[:idx]
}
func (t *Table) tail(p *Piece, idx int) string {
	return t.source(p.Source)[idx:]
}

func (t *Table) allContents() string {
	//return t.contents(0, t.size())
	// need golang's stringbuilder
	s := ""

	t.dump()
	for i := 0; i < len(t.Mods); i++ {
		s += t.runForMod(i)
	}
	log.Println("ac: ", s)
	return string(s)
}

// need to add observance of start, end
func (t *Table) contents(start, end int) string {
	// need golang's stringbuilder
	s := ""
	si, ss := t.pieceAt(start)
	sp := t.Mods[si]
	ei, ee := t.pieceAt(end)
	ep := t.Mods[ei]
	startFrag := t.head(sp, ss)
	endFrag := t.tail(ep, ee)
	s += startFrag
	for i := si + 1; i < ei; i++ {
		s += t.runForMod(i)
	}
	s += endFrag
	return s
}

func (t *Table) indexOf(idx int) byte {
	e, i := t.pieceAt(idx)
	return t.runForMod(e)[i]
}

func (t *Table) appendPiece(p *Piece) {
	t.insertPieceAt(len(t.Mods), p)
}

func (t *Table) deletePieceAt(index int) {
	t.Mods = append(t.Mods[:index], t.Mods[index+1:]...)
}

func (t *Table) insertPieceAt(index int, p *Piece) {
	if index >= len(t.Mods) {
		t.Mods = append(t.Mods, p)
		return
	}
	t.Mods = append(t.Mods[:index+1], t.Mods[index:]...)
	t.Mods[index] = p
	t.dump()
	return
}

func (t *Table) pieceAt(idx int) (int, int) {
	i := idx
	for j, p := range t.Mods {
		if i <= p.Run {
			return j, i
		} else {
			i = i - p.Run
		}
	}
	return 0, 0
}

func (t *Table) add(s string, pt int) error {
	e, i := t.pieceAt(pt)
	p := t.Mods[e]
	// Appending characters to the "add file" buffer, and
	np := NewPiece(Add, len(t.Add), len(s))
	t.Add += s
	np.dump(log.New(os.Stderr, "np ", 0))
	// Updating the entry in piece table (breaking an entry into two or three)
	if i == 0 {
		// insert at p
		t.insertPieceAt(e, np)
		return nil
	}
	if i == p.Run {
		// insert np at p+1
		//t.insertPieceAt(e+1, np)
		t.appendPiece(np)
		return nil
	}
	// else split the piece and make { left, np, right }
	left, right := p.splitAt(i)
	t.deletePieceAt(e)
	t.insertPieceAt(e, left)
	t.insertPieceAt(e+1, right)
	t.insertPieceAt(e+1, np)
	return nil
}

func (t *Table) deleteRune(idx int) {
	which, i := t.pieceAt(idx)
	p := t.Mods[which]
	if i == 0 || i == p.Run {
		p.trimRune(i)
	}
	// else split into two pieces
	left, right := p.splitAt(i)
	//left.Run -= 1 // trim last rune of left (orig idx above)
	right.Start += 1
	right.Run -= 1
	t.deletePieceAt(which)
	t.insertPieceAt(which, right)
	t.insertPieceAt(which, left)
}

// load a text file from a filename string
func LoadFile(filename string) *Table {
	content, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Fatal(err)
	}

	return NewTable(string(content))
}

func (t *Table) SaveToFile(filename string) error {
	s := t.allContents()
	f, err := os.Create("test.txt")
	if err != nil {
		return err
	}
	_, err = f.WriteString(s)
	if err != nil {
		f.Close()
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}
