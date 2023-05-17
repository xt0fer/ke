package buffer

import (
	"log"
	"os"
	"runtime"
)

func (t *Table) dump() {
	l := log.New(os.Stderr, "", 0)
	_, filename, line, _ := runtime.Caller(1)
	l.Println(">> Table Dump", "@", filename, line)

	l.Println("Content", &t.Content, t.Content)
	l.Println("Add    ", &t.Add, t.Add)
	for i := 0; i < len(t.Mods); i++ {
		p := t.Mods[i]
		p.dump(l)
		l.Println(p, t.runForMod(i))
	}
	l.Println(">> End")

}

func (p *Piece) dump(f *log.Logger) {
	_, filename, line, _ := runtime.Caller(1)
	f.Println(p.Source, p.Start, p.Run, "->", filename, line)
}
