package kg

import (
	"fmt"
	"log"
	"strings"
	"unicode"

	"github.com/gorilla/websocket"
	"github.com/kristofer/ke/term"
	// "github.com/nsf/termbox-go"
)

func checkErr(e error) {
	if e != nil {
		panic(e)
	}
}

const (
	version        = "kg 2.0, Public Domain, May 2023, Kristofer Younger,  No warranty."
	nomark         = -1
	gapchunk       = 16 //= 8096
	idDefault      = 1
	idSymbol       = 2
	idModeline     = 3
	idDigits       = 4
	idLineComment  = 5
	idBlockComment = 6
	idDoubleString = 7
	idSingleString = 8
	initialText    = "1 foo bar baz\n2 foo baz kristofer\n3 hello there.\n1234567890123456789012345678901234567890\n"
)

// Editor struct
type Editor struct {
	Term      *term.Term
	InputChan chan term.Event
	//EventChan     chan termbox.Event
	CurrentBuffer *Buffer /* current buffer */
	RootBuffer    *Buffer /* head of list of buffers */
	CurrentWindow *Window
	RootWindow    *Window
	// status vars
	Done          bool   /* Quit flag. */
	Msgflag       bool   /* True if msgline should be displayed. */
	PasteBuffer   string /* Allocated scrap buffer. */
	Msgline       string /* Message line input/output buffer. */
	Searchtext    string
	Replace       string
	Keymap        []keymapt
	Lines         int
	Cols          int
	FGColor       term.Attribute
	BGColor       term.Attribute
	EscapeFlag    bool
	CtrlXFlag     bool
	MiniBufActive bool
}

// StartEditor is the old C main function
func (e *Editor) StartEditor(argv []string, argc int, conn *websocket.Conn) {
	// log setup....
	// f, err := os.OpenFile("logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// if err != nil {
	// 	log.Fatalf("error opening file: %v", err)
	// }
	// defer f.Close()

	// log.SetOutput(f)
	// f.Truncate(0)
	// log.Println("Start of Log...")
	//
	// e.FGColor = termbox.ColorDefault
	// e.BGColor = termbox.ColorWhite
	// err = termbox.Init()
	//checkErr(err)
	// defer termbox.Close()
	// e.Cols, e.Lines = termbox.Size()
	e.Term = term.NewTerm(term.Web)

	e.Term.Kind = term.Web
	e.Term.Conn = conn
	e.Cols, e.Lines = e.Term.Size()

	//editor.msg("NO file to open, creating scratch buffer")
	e.CurrentBuffer = e.FindBuffer("*scratch*", true)
	e.CurrentBuffer.Buffername = "*scratch*"
	e.CurrentBuffer.setText(initialText)
	//editor.top()

	e.CurrentWindow = NewWindow(e)
	e.RootWindow = e.CurrentWindow
	e.CurrentWindow.OneWindow()
	e.CurrentWindow.AssociateBuffer(e.CurrentBuffer)

	if !(e.CurrentBuffer.GrowGap(16)) {
		panic("%s: Failed to allocate required memory.\n")
	}
	e.Keymap = Keymap

	//m :=
	e.UpdateDisplay()
	e.Term.Flush()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("unable to get message from frontend")
			return
		}
		//log.Printf("ev: |%x| |%s| \n", msg, string(msg))

		event := e.Term.EventFromKey(msg)

		ok := e.HandleEvent(&event)
		if !ok {
			conn.Close()
			break //exit editor
		}

		e.UpdateDisplay()

		e.Term.Flush()
	}
}

// handleEvent
func (e *Editor) HandleEvent(ev *term.Event) bool {
	e.msg("")
	switch ev.Type {
	case term.EventKey:
		if ev.Ch != 0 && (e.CtrlXFlag || e.EscapeFlag) {
			_ = e.OnSysKey(ev)
			if e.Done {
				return false
			}
		} else if ev.Ch == 0 {
			_ = e.OnSysKey(ev)
			if e.Done {
				return false
			}
		} else {
			//log.Println("e.CurrentWindow.OnKey", ev.String())
			e.CurrentWindow.OnKey(ev)
		}
		e.UpdateDisplay()
	case term.EventResize:
		e.Term.Clear()
		e.Cols, e.Lines = e.Term.Size()
		e.msg("Resize: h %d,w %d", e.Lines, e.Cols)
		e.CurrentWindow.WindowResize()
		e.UpdateDisplay()
	case term.EventMouse:
		e.Term.Clear()
		e.msg("Mouse: r %d, c %d ", ev.MouseY, ev.MouseX)
		e.SetPointForMouse(ev.MouseX, ev.MouseY)
		e.UpdateDisplay()
	case term.EventError:
		panic(ev.Err)
	}

	return true
}

// OnSysKey on Ctrl key pressed
func (e *Editor) OnSysKey(ev *term.Event) bool {
	switch ev.Key {
	case term.KeyCtrlX:
		log.Println("C-X")
		e.msg("C-X ")
		e.CtrlXFlag = true
		return true
	case term.KeyEsc:
		e.msg("Esc ")
		e.EscapeFlag = true
		return true
	case term.KeyCtrlQ:
		e.Done = true
		return true
	case term.KeySpace, term.KeyEnter, term.KeyCtrlJ, term.KeyTab:
		e.CurrentWindow.OnKey(ev)
		return true
	case term.KeyArrowDown, term.KeyArrowLeft, term.KeyArrowRight, term.KeyArrowUp:
		e.CtrlXFlag = false
		e.EscapeFlag = false
		return e.RunKeymapFunction(ev)
	default:
		return e.RunKeymapFunction(ev)
	}
}

func (e *Editor) RunKeymapFunction(ev *term.Event) bool {
	rch := ev.Ch
	if ev.Ch == 0 {
		rch = rune(ev.Key)
	}
	lookfor := fmt.Sprintf("%c", rch)
	if e.CtrlXFlag {
		lookfor = fmt.Sprintf("\x18%c", rch)
	}
	if e.EscapeFlag {
		lookfor = fmt.Sprintf("\x1B%c", rch)
	}
	for i, j := range e.Keymap {
		if strings.Compare(lookfor, j.KeyBytes) == 0 {
			//log.Println("SearchAndPerform FOUND ", lookfor, e.Keymap[i])
			do := e.Keymap[i].Do
			if do != nil {
				do(e) // execute function for key
			}
			e.CtrlXFlag = false
			e.EscapeFlag = false
			return true
		}
	}
	return false
}

// OnAltKey on Alt key pressed
func (e *Editor) OnAltKey(ev *term.Event) bool {
	e.msg("AltKey")
	return false
}

func (e *Editor) msg(fm string, args ...interface{}) {
	e.Msgline = fmt.Sprintf(fm, args...)
	e.Msgflag = true
}

func (e *Editor) drawString(x, y int, fg, bg term.Attribute, msg string) {
	for _, c := range msg {
		e.Term.SetCell(x, y, c, fg, bg)
		x++
	}
}

func (e *Editor) displayMsg() {
	if e.Msgflag {
		e.drawString(0, e.Lines-1, e.FGColor, term.ColorDefault, e.Msgline)
	}
	e.blankFrom(e.Lines-1, len(e.Msgline))
}

// Display draws the window, minding the buffer pagestart/pageend
func (e *Editor) Display(wp *Window, shouldDrawCursor bool) {
	e.Term.Blank()
	//	e.Term.SetCursor(0, 0)
	bp := wp.Buffer
	pt := bp.Point
	// /* find start of screen, handle scroll up off page or top of file  */
	if pt < bp.PageStart {
		bp.PageStart = bp.SegStart(bp.LineStart(pt), pt, e.Cols)
	}

	if bp.Reframe || (pt > bp.PageEnd && pt != bp.PageEnd && !(pt >= bp.TextSize)) {
		bp.Reframe = false
		i := 0
		/* Find end of screen plus one. */
		bp.PageStart = bp.DownDown(pt, e.Cols)
		/* if we scroll to EOF we show 1 blank line at bottom of screen */
		if bp.PageEnd <= bp.PageStart {
			bp.PageStart = bp.PageEnd
			i = wp.Rows - 1 // 1
		} else {
			i = wp.Rows - 0
		}
		/* Scan backwards the required number of lines. */
		for i > 0 {
			bp.PageStart = bp.UpUp(bp.PageStart, e.Cols)
			i--
		}
	}

	l1 := bp.LineForPoint(bp.PageStart)
	l2 := l1 + wp.Rows
	l2end := bp.LineEnd(bp.PointForLine(l2))
	bp.PageEnd = l2end
	r, c := wp.TopPt, 0
	for k := bp.PageStart; k <= bp.PageEnd; k++ {
		if pt == k {
			bp.PointCol = c
			bp.PointRow = r
		}
		rch, err := bp.RuneAt(k)
		if err != nil {
			e.msg("Error on RuneAt", err)
		}
		if rch != '\r' {
			if unicode.IsPrint(rch) || rch == '\t' || rch == '\n' {
				if rch == '\t' {
					c += 3 //? 8-(j&7) : 1;
				}
				if rch != '\n' {
					e.Term.SetCell(c, r, rch, e.FGColor, term.ColorDefault)
					c++
				} else {
					//log.Println("found a newline,", r)
				}
			} else {
				e.Term.SetCell(c, r, rch, e.FGColor, term.ColorDefault)
				c++
			}
		}

		if rch == '\n' || e.Cols <= c {
			//log.Println("displaying NewLine", c, r)
			e.blankFrom(r, c)
			c -= e.Cols
			if c < 0 {
				c = 0
			}
			r++
		}
	}
	for k := r; k < wp.TopPt+wp.Rows+1; k++ {
		e.blankFrom(k, 0)
	}

	buffer2Window(wp)
	e.ModeLine(wp)
	if wp == e.CurrentWindow && shouldDrawCursor {
		e.displayMsg()
		e.setTermCursor(wp.Col, wp.Row) //bp.PointCol, bp.PointRow)
	}
	//term.Flush() //refresh();
	wp.Updated = false
}

func (e *Editor) blankFrom(r, c int) { // blank line to end of term
	// hmm. deep bug? why e.cols -1 ??
	ch := ' '
	for k := c; k < e.Cols; k++ {
		ch = '~'
		if k == 0 {
			ch = '^'
		}
		if k == e.Cols-1 {
			ch = '$'
		}
		e.Term.SetCell(k, r, ch, e.FGColor, term.ColorDefault)
	}
}
func (e *Editor) setTermCursor(c, r int) {
	//log.Printf("editor setTermCursor %d, %d\n", c, r)
	wp := e.CurrentWindow
	wp.Col, wp.Row = c, r
	e.Term.SetCursor(c, r)
}

func (e *Editor) UpdateDisplay() {
	bp := e.CurrentWindow.Buffer
	bp.OrigPoint = bp.Point /* OrigPoint only ever set here */
	/* only one window */
	if e.RootWindow.Next == nil {
		e.Display(e.CurrentWindow, true)
		//		term.Flush()
		bp.PrevSize = bp.TextSize
		return
	}
	/* this is key, we must call our win first to get accurate page and epage etc */
	e.Display(e.CurrentWindow, false)
	/* never CurrentWin,  but same buffer in different window or update flag set*/
	for wp := e.RootWindow; wp != nil; wp = wp.Next {
		if wp != e.CurrentWindow && (wp.Buffer == bp || wp.Updated) {
			window2Buffer(wp)
			e.Display(wp, false)
		}
	}
	/* now display our window and buffer */
	window2Buffer(e.CurrentWindow)
	e.displayMsg()
	e.setTermCursor(e.CurrentWindow.Col, e.CurrentWindow.Row)
	bp.PrevSize = bp.TextSize /* now safe to save previous size for next time */
}

// SetPointForMouse xxx
func (e *Editor) SetPointForMouse(mc, mr int) {
	c, r := e.setWindowForMouse(mc, mr)
	bp := e.CurrentBuffer
	sl := bp.LineForPoint(bp.PageStart) // sl is startline of buffer frame
	ml := sl + r
	mlpt := bp.PointForLine(ml)
	mll := bp.LineLenAtPoint(mlpt) // how wide is line?
	nc := c + 1
	if mll < c {
		nc = mll
	}
	npt := bp.PointForXY(nc, ml)
	bp.SetPoint(npt)
}

func (e *Editor) setWindowForMouse(mc, mr int) (c, r int) {
	log.Printf("setWindowForMouse col %d row %d ", mc, mr)

	wp := e.RootWindow
	// if mr is modeline or modeline+1, reduce to last wp.Rows
	if mr > wp.Rows {
		mr = wp.Rows
	}
	for wp != nil {
		if (mr <= wp.Rows+wp.TopPt) && (mr >= wp.TopPt) {
			log.Printf("set win rows %d top %d\n", wp.Rows, wp.TopPt)
			e.setWindow(wp)
			r = mr - wp.TopPt
			// if mr == wp.Rows+wp.TopPt {
			// 	r--
			// }
			c = mc
			return
		}
		wp = wp.Next
	}
	return 0, e.Lines - 1
}

// ModeLine draw modeline for window
func (e *Editor) ModeLine(wp *Window) {
	var lch, mch, och rune
	e.Cols, e.Lines = e.Term.Size()

	if wp == e.CurrentWindow {
		lch = '='
	} else {
		lch = '-'
	}
	mch = lch
	if wp.Buffer.modified {
		mch = '*'
	}
	och = lch
	temp := fmt.Sprintf("%c%c%c kg: %c%c %s L%d wp(%d,%d)", lch, och, mch, lch, lch,
		e.GetBufferName(wp.Buffer),
		wp.Buffer.PointRow, wp.Row, wp.Col)
	x := 0
	y := wp.TopPt + wp.Rows + 1
	for _, c := range temp {
		e.Term.SetCell(x, y, c, term.ColorBlack, e.BGColor)
		x++
	}

	for i := len(temp); i <= e.Cols; i++ {
		e.Term.SetCell(i, y, lch, term.ColorBlack, e.BGColor) // e.FGColor
	}
}

func (e *Editor) displayPromptAndResponse(prompt string, response string) {
	e.drawString(0, e.Lines-1, e.FGColor, term.ColorDefault, prompt)
	if response != "" {
		e.drawString(len(prompt), e.Lines-1, e.FGColor, term.ColorDefault, response)
	}
	e.blankFrom(e.Lines-1, len(prompt)+len(response))
	e.Term.SetCursor(len(prompt)+len(response), e.Lines-1)
	// term.Flush()
}

func (e *Editor) GetInput(prompt string) string {
	fname := ""
	var ev term.Event
	e.displayPromptAndResponse(prompt, "")
	e.MiniBufActive = true
loop:
	for {
		ev = <-e.InputChan
		if ev.Ch != 0 {
			ch := ev.Ch
			fname = fname + string(ch)
		}
		if ev.Ch == 0 {
			switch ev.Key {
			case term.KeyTab:
				fname = fname + string('\t')
			case term.KeySpace:
				fname = fname + string(' ')
			case term.KeyEnter, term.KeyCtrlR:
				break loop
			case term.KeyBackspace2, term.KeyBackspace:
				if len(fname) > 0 {
					fname = fname[:len(fname)-1]
				} else {
					fname = ""
				}
			case term.KeyCtrlG:
				return ""
			default:

			}
		}
		e.displayPromptAndResponse(prompt, fname)
	}
	e.MiniBufActive = false
	return fname
}

// DeleteBuffer unlink from the list of buffers, free associated memory,
// assumes buffer has been saved if modified
func (e *Editor) deleteBuffer(bp *Buffer) bool {
	var sb *Buffer

	/* we must have switched to a different buffer first */
	if bp != e.CurrentBuffer {
		/* if buffer is the head buffer */
		if bp == e.RootBuffer {
			e.RootBuffer = bp.Next
		} else {
			/* find place where the bp buffer is next */
			for sb = e.RootBuffer; sb.Next != bp && sb.Next != nil; sb = sb.Next {
			}
			if sb.Next == bp || sb.Next == nil {
				sb.Next = bp.Next
			}
		}
		bp = nil
	} else {
		return false
	}
	return true
}

// NextBuffer returns next buffer after current
func (e *Editor) nextBuffer() {
	if e.CurrentBuffer != nil && e.RootBuffer != nil {
		e.CurrentWindow.DisassociateBuffer()
		if e.CurrentBuffer.Next != nil {
			e.CurrentBuffer = e.CurrentBuffer.Next

		} else {
			e.CurrentBuffer = e.RootBuffer
		}
		e.CurrentWindow.AssociateBuffer(e.CurrentBuffer)
		e.CurrentBuffer.Reframe = true
	}
}

// GetBufferName returns buffer name
func (e *Editor) GetBufferName(bp *Buffer) string {
	if bp.Filename != "" {
		return bp.Filename
	}
	return bp.Buffername
}

// CountBuffers how many buffers in list
func (e *Editor) CountBuffers() int {
	var bp *Buffer
	i := 0

	for bp = e.RootBuffer; bp != nil; bp = bp.Next {
		i++
	}
	return i
}

// ModifiedBuffers true is any buffers modified
func (e *Editor) ModifiedBuffers() bool {
	var bp *Buffer

	for bp = e.RootBuffer; bp != nil; bp = bp.Next {
		if bp.modified {
			return true
		}
	}
	return false
}

// FindBuffer Find a buffer by filename or create if requested
func (e *Editor) FindBuffer(fname string, cflag bool) *Buffer {
	bp := e.RootBuffer
	for bp != nil {
		if strings.Compare(fname, bp.Filename) == 0 || strings.Compare(fname, bp.Buffername) == 0 {
			return bp
		}
		bp = bp.Next
	}
	if cflag {
		bp = NewBuffer()
		/* find the place in the list to insert this buffer */
		if e.RootBuffer == nil {
			e.RootBuffer = bp
		} else if strings.Compare(e.RootBuffer.Filename, fname) > 0 {
			/* insert at the begining */
			bp.Next = e.RootBuffer
			e.RootBuffer = bp
		} else {
			sb := e.RootBuffer
			for sb.Next != nil {
				if strings.Compare(sb.Next.Filename, fname) > 0 {
					break
				}
				sb = sb.Next
			}
			/* and insert it */
			bp.Next = sb.Next
			sb.Next = bp
		}
	}
	return bp
}

func (e *Editor) splitWindow() {
	if e.CurrentWindow.Rows < 3 {
		e.msg("Cannot split a %d line window", e.CurrentWindow.Rows)
		return
	}

	nwp := NewWindow(e)
	nwp.AssociateBuffer(e.CurrentWindow.Buffer)
	buffer2Window(nwp)

	ntru := (e.CurrentWindow.Rows - 1) / 2    /* Upper size */
	ntrl := (e.CurrentWindow.Rows - 1) - ntru /* Lower size */

	/* Old is upper window */
	e.CurrentWindow.Rows = ntru
	nwp.TopPt = e.CurrentWindow.TopPt + ntru + 2
	nwp.Rows = ntrl - 1

	/* insert it in the list */
	wp2 := e.CurrentWindow.Next
	e.CurrentWindow.Next = nwp
	nwp.Next = wp2
	/* mark the lot for update */
	e.redraw()
}

// NextWindow
func (e *Editor) nextWindow() {
	e.CurrentWindow.Updated = true /* make sure modeline gets updated */
	//Curwp = (Curwp.Next == nil ? Wheadp : Curwp.Next)
	if e.CurrentWindow.Next == nil {
		e.CurrentWindow = e.RootWindow
	} else {
		e.CurrentWindow = e.CurrentWindow.Next
	}
	e.CurrentBuffer = e.CurrentWindow.Buffer

	if e.CurrentBuffer.WinCount > 1 {
		/* push win vars to buffer */
		window2Buffer(e.CurrentWindow)
	}
}

func (e *Editor) setWindow(wp *Window) {
	e.CurrentWindow.Updated = true /* make sure modeline gets updated */
	e.CurrentWindow = wp
	e.CurrentWindow.Updated = true /* make sure modeline gets updated */
	e.CurrentBuffer = e.CurrentWindow.Buffer
	// if e.CurrentBuffer.WinCount > 1 {
	// 	/* push win vars to buffer */
	// 	window2Buffer(e.CurrentWindow)
	// }
	e.UpdateDisplay()
}

// DeleteOtherWindows
func (e *Editor) deleteOtherWindows() {
	wp := e.RootWindow
	if wp.Next == nil {
		wp.Editor.msg("Only 1 window")
		return
	}
	e.freeOtherWindows()
}

// FreeOtherWindows
func (e *Editor) freeOtherWindows() {
	wp := e.RootWindow
	winp := e.CurrentWindow
	next := wp
	for next != nil {
		next = wp.Next /* get next before a call to free() makes wp undefined */
		if wp != winp {
			wp.DisassociateBuffer() /* this window no longer references its buffer */
		}
		wp = next
	}
	e.RootWindow = winp
	e.CurrentWindow = winp
	winp.OneWindow()
}
