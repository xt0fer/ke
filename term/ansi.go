package term

import "log"

// command.go:     termbox "github.com/nsf/termbox-go"
// command.go:     termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
func (t *Term) Clear() {
	//termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	if t.IsPty() {
		t.Output.Write([]byte(ED(2)))
	}
	if t.IsWeb() {
		t.Write([]byte(ED(2)))
		//t.Write([]byte(CURBLK()))
	}
}

// editor.go:      termbox "github.com/nsf/termbox-go"
// editor.go:      EventChan     chan termbox.Event
// editor.go:      FGColor       termbox.Attribute
// editor.go:      BGColor       termbox.Attribute
// editor.go:      e.FGColor = termbox.ColorDefault
// editor.go:      e.BGColor = termbox.ColorWhite
// editor.go:      err = termbox.Init()
// editor.go:      defer termbox.Close()
// editor.go:      e.Cols, e.Lines = termbox.Size()

func (t *Term) Size() (int, int) {
	return 40, 12 // termsize cols, rows; 80, 24
}

// editor.go:      termbox.SetInputMode(termbox.InputAlt | termbox.InputEsc | termbox.InputMouse)
// editor.go:      termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
// editor.go:      termbox.Flush()
// editor.go:      e.EventChan = make(chan termbox.Event, 20)
// editor.go:                      e.EventChan <- termbox.PollEvent()
// editor.go:              termbox.Flush()
// editor.go:func (e *Editor) handleEvent(ev *termbox.Event) bool {
// editor.go:      case termbox.EventKey:
// editor.go:              if (ev.Mod & termbox.ModAlt) != 0 {
// editor.go:                      // if ev.Mod&termbox.ModAlt != 0 && e.OnAltKey(ev) {
// editor.go:                      if (ev.Mod & termbox.ModAlt) != 0 {
// editor.go:      case termbox.EventResize:
// editor.go:              termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
// editor.go:              e.Cols, e.Lines = termbox.Size()
// editor.go:      case termbox.EventMouse:
// editor.go:              termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
// editor.go:      case termbox.EventError:
// editor.go:func (e *Editor) OnSysKey(ev *termbox.Event) bool {
// editor.go:      case termbox.KeyCtrlX:
// editor.go:      case termbox.KeyEsc:
// editor.go:      case termbox.KeyCtrlQ:
// editor.go:      case termbox.KeySpace, termbox.KeyEnter, termbox.KeyCtrlJ, termbox.KeyTab:
// editor.go:      case termbox.KeyArrowDown, termbox.KeyArrowLeft, termbox.KeyArrowRight, termbox.KeyArrowUp:
// editor.go:func (e *Editor) searchAndPerform(ev *termbox.Event) bool {
// editor.go:func (e *Editor) OnAltKey(ev *termbox.Event) bool {
// editor.go:func (e *Editor) drawString(x, y int, fg, bg termbox.Attribute, msg string) {
// editor.go:              termbox.SetCell(x, y, c, fg, bg)
// editor.go:              e.drawString(0, e.Lines-1, e.FGColor, termbox.ColorDefault, e.Msgline)
// editor.go:                              termbox.SetCell(c, r, rch, e.FGColor, termbox.ColorDefault)
// editor.go:                              termbox.SetCell(c, r, rch, e.FGColor, termbox.ColorDefault)
// editor.go:      termbox.Flush() //refresh();
// editor.go:              termbox.SetCell(k, r, ' ', e.FGColor, termbox.ColorDefault)
// editor.go:      termbox.SetCursor(c, r)
// editor.go:              termbox.Flush()
// editor.go:      e.Cols, e.Lines = termbox.Size()
// editor.go:              termbox.SetCell(x, y, c, termbox.ColorBlack, e.BGColor)
// editor.go:              termbox.SetCell(i, y, lch, termbox.ColorBlack, e.BGColor) // e.FGColor
func (t *Term) SetCell(c, r int, ch rune, fg, bg Attribute) {
	// switch zero-based to one-based?
	t.ScrBuf.Set(c, r, ch)
}
func (t *Term) SetCursor(c int, r int) {
	log.Println("cursor", c, r)
	t.CurCol = c
	t.CurRow = r
	//^[[<r>;<c>H
	if t.IsWeb() {
		t.Write([]byte(CUP(c, r)))
	}
}

// editor.go:      e.drawString(0, e.Lines-1, e.FGColor, termbox.ColorDefault, prompt)
// editor.go:              e.drawString(len(prompt), e.Lines-1, e.FGColor, termbox.ColorDefault, response)
// editor.go:      termbox.SetCursor(len(prompt)+len(response), e.Lines-1)
// editor.go:      termbox.Flush()
// editor.go:      var ev termbox.Event
// editor.go:                      case termbox.KeyTab:
// editor.go:                      case termbox.KeySpace:
// editor.go:                      case termbox.KeyEnter, termbox.KeyCtrlR:
// editor.go:                      case termbox.KeyBackspace2, termbox.KeyBackspace:
// editor.go:                      case termbox.KeyCtrlG:

// made changes in window.go

// window.go:      termbox "github.com/nsf/termbox-go"
// window.go:      //termbox "github.com/gdamore/tcell/termbox"
// window.go:func (wp *Window) OnKey(ev *termbox.Event) {
// window.go:      case termbox.KeySpace:
// window.go:      case termbox.KeyEnter, termbox.KeyCtrlJ:
// window.go:      case termbox.KeyTab:
// window.go:              if ev.Mod&termbox.ModAlt != 0 && wp.Editor.OnAltKey(ev) {
