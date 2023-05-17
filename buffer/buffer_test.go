package buffer

import (
	"fmt"
	"testing"
)

func TestSomething(t *testing.T) {
	s := "this is a test string."

	if len(s) <= 0 {
		t.Errorf("len() doesn't work!?")
	}
}

func TestPieceSplit1(t *testing.T) {
	s := "this is a test string."
	//    0123456789123456789212
	p := NewPiece(s, 0, len(s))

	if p.size() != len(s) {
		t.Errorf("Piece isn't the right size")
	}

	at := 1
	left, right := p.splitAt(at)

	if left.size() != at && (p.size()-right.size()) != len(s)-at {
		t.Errorf("p.splitAt(5) doesn't work!?")
		t.Errorf("%v", p)
		t.Errorf("%v, %v", left, right)
	}

}

func TestPieceSplit2(t *testing.T) {
	s := "this is a test string."
	//    0123456789123456789212
	p := NewPiece(s, 0, len(s))

	if p.size() != len(s) {
		t.Errorf("Piece isn't the right size")
	}

	at := 5
	left, right := p.splitAt(at)

	if left.size() != at && (p.size()-right.size()) != len(s)-at {
		t.Errorf("p.splitAt(1) doesn't work!?")
		t.Errorf("%v", p)
		t.Errorf("%v, %v", left, right)
	}
}

func TestPieceSplit3(t *testing.T) {
	s := "this is a test string."
	//    0123456789123456789212
	p := NewPiece(s, 0, len(s))

	if p.size() != len(s) {
		t.Errorf("Piece isn't the right size")
	}

	at := 16
	left, right := p.splitAt(at)

	if left.size() != at && (p.size()-right.size()) != len(s)-at {
		t.Errorf("p.splitAt(1) doesn't work!?")
		t.Errorf("%v", p)
		t.Errorf("%v, %v", left, right)
	}
}

func TestLoadFile(t *testing.T) {
	fname := "testtext.txt"

	tt := LoadFile(fname)
	if tt.size() <= 0 {
		t.Errorf("File Size is bad")
	}

}

func TestSize1(t *testing.T) {
	//fname := "testtext.txt"
	s := "this is a test. "

	tt := NewTable(s)

	b := tt.size()
	if b != len(s) {
		t.Errorf("TestSize1 failed.")
	}
}

// func TestAddToBuffer(t *testing.T) {
// 	fname := "testtext.txt"
// 	s := "this is a test. "

// 	tt := LoadFile(fname)
// 	if tt.size() <= 0 {
// 		t.Errorf("File Size is bad")
// 	}

// 	err := tt.add(s, 5)
// 	if err != nil {
// 		t.Errorf("Add failed.")
// 	}

// }

func TestIndex1(t *testing.T) {
	//fname := "testtext.txt"
	s := "this is a test. "

	tt := NewTable(s)

	b := tt.index(0)
	if b != 't' {
		t.Errorf("Index failed.")
	}
	//t.Errorf("b is %s", string(b))
	b1 := tt.index(11)
	if b1 != 'e' {
		t.Errorf("Index failed.")
	}
	//t.Errorf("b is %s", string(b))
	b2 := tt.index(5)
	if b2 != 'i' {
		t.Errorf("Index failed.")
	}
	//t.Errorf("b is %s", string(b))
}

func TestInsertPiece(t *testing.T) {
	//fname := "testtext.txt"
	s := "this is a test. "

	tt := NewTable(s)

	foo := "foo"
	p := NewPiece(foo, 0, len(foo))
	tt.insertPieceAt(0, p)
	foo2 := "foo2"
	q := NewPiece(foo2, 0, len(foo2))
	tt.insertPieceAt(0, q)

	for i, p := range tt.Mods {
		fmt.Printf("%+v %+v\n", i, p)
	}
	t.Errorf("tt.Mods %+v", tt.Mods)
}

func TestAppendPiece(t *testing.T) {
	//fname := "testtext.txt"
	s := "this is a test. "

	tt := NewTable(s)

	foo := "foo"
	p := NewPiece(foo, 0, len(foo))
	tt.appendPiece(p)
	foo2 := "foo2"
	q := NewPiece(foo2, 0, len(foo2))
	tt.appendPiece(q)

	for i, p := range tt.Mods {
		fmt.Printf("%+v %+v\n", i, p)
	}
	t.Errorf("tt.Mods %+v", tt.Mods)
}
func TestAppendInsertPiece(t *testing.T) {
	//fname := "testtext.txt"
	s := "this is a test. "

	tt := NewTable(s)

	foo := "foo"
	p := NewPiece(foo, 0, len(foo))
	tt.appendPiece(p)
	foo2 := "foo2"
	q := NewPiece(foo2, 0, len(foo2))
	tt.insertPieceAt(1, q)

	for i, p := range tt.Mods {
		fmt.Printf("%+v %+v\n", i, p)
	}
	t.Errorf("tt.Mods %+v", tt.Mods)
}

func TestAllContents(t *testing.T) {
	//fname := "testtext.txt"
	s := "this is a test. "

	tt := NewTable(s)

	s0 := tt.allContents()
	if s != s0 {
		t.Errorf("s0 %+v", s0)
	}
}
