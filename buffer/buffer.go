package buffer

import (
	"io/ioutil"
	"log"
)

type Table struct {
	Content string
	Add     string
	Mods    []Piece
}

type Piece struct {
	Source string
	Start  int
	Run    int
}

func NewTable(c string) *Table {
	return &Table{Content: c, Add: "", Mods: []Piece{}}
}

func (t Table) size() int {
	return len(t.Content)
}

// load a text file from a filename string
func LoadFile(filename string) *Table {
	content, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Fatal(err)
	}

	return NewTable(string(content))
}
