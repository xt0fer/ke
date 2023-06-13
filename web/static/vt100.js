// VT100.js -- a text terminal emulator in JavaScript with a ncurses-like
// interface and a POSIX-like interface. (The POSIX-like calls are
// implemented on top of the ncurses-like calls, not the other way round.)
//
// Released under the GNU LGPL v2.1, by Frank Bi <bi@zompower.tk>
//Added to by Kristofer Younger <kris atz zipcodewilmington dot com>
// 2023-06-01 - upgraded to osrta latest JS
// 2007-08-12	- refresh():
//		  - factor out colour code to html_colours_()
//		  - fix handling of A_REVERSE | A_DIM
//		  - simplify initial <br /> output code
//		  - fix underlining colour
//		- fix attron() not to turn off attributes
//		- decouple A_STANDOUT and A_BOLD
// 2007-08-11	- getch() now calls refresh()
// 2007-08-06	- Safari compat fix -- turn '\r' into '\n' for onkeypress
// 2007-08-05	- Opera compat fixes for onkeypress
// 2007-07-30	- IE compat fixes:
//		  - change key handling code
//		  - add <br />...<br />&nbsp; so that 1st and last lines align
// 2007-07-28	- change wrapping behaviour -- writing at the right edge no
//		  longer causes the cursor to immediately wrap around
//		- add <b>...</b> to output to make A_STANDOUT stand out more
//		- add handling of backspace, tab, return keys
//		- fix doc. of VT100() constructor
//		- change from GPL to LGPL
// 2007-07-09	- initial release
//
// class VT100
//	A_NORMAL, A_UNDERLINE, A_REVERSE, A_BLINK, A_DIM, A_BOLD, A_STANDOUT
//	=class constants=
//			Attribute constants.
//	VT100(wd, ht, scr_id) =constructor=
//			Creates a virtual terminal with width `wd', and
//			height `ht'. The terminal will be displayed between
//			<pre>...</pre> tags which have element ID `scr_id'.
//	addch(ch [, attr])
//			Writes out the character `ch'. If `attr' is given,
//			it specifies the attributes for the character,
//			otherwise the current attributes are used.
//	addstr(stuff)	Writes out the string `stuff' using the current
//			attributes.
//	attroff(mode)	Turns off any current options given in mode.
//	attron(mode)	Turns on any options given in mode.
//	attrset(mode)	Sets the current options to mode.
//	bkgdset(attr)	Sets the background attributes to attr.
//	clear()		Clears the terminal using the background attributes,
//			and homes the cursor.
//	clrtobol()	Clears the portion of the terminal from the cursor
//			to the bottom.
//	clrtoeol()	Clears the portion of the current line after the
//			cursor.
//	curs_set(vis [, grab])
//			If `vis' is 0, makes the cursor invisible; otherwise
//			make it visible. If `grab' is given and true, starts
//			capturing keyboard events (for `getch()'); if given
//			and false, stops capturing events.
//	echo()		Causes key strokes to be automatically echoed on the
//			terminal.
//	erase()		Same as `clear()'.
//	getch(isr)	Arranges to call `isr' when a key stroke is
//			received. The received character and the terminal
//			object are passed as arguments to `isr'.
//	getmaxyx()	Returns an associative array with the maximum row
//			(`y') and column (`x') numbers for the terminal.
//	getyx()		Returns an associative array with the current row
//			(`y') and column (`x') of the cursor.
//	move(r, c)	Moves the cursor to row `r', column `c'.
//	noecho()	Stops automatically echoing key strokes.
//	refresh()	Updates the display.
//	scroll()	Scrolls the terminal up one line.
//	standend()	Same as `attrset(VT100.A_NORMAL)'.
//	standout()	Same as `attron(VT100.A_STANDOUT)'.
//	write(stuff)	Writes `stuff' to the terminal and immediately
//			updates the display; (some) escape sequences are
//			interpreted and acted on.

// constructor
function VT100(wd, ht, scr_id) {
    var r;
    var c;
    var scr = document.getElementById(scr_id);
    this.wd_ = wd;
    this.ht_ = ht;
    this.scrolled_ = 0;
    this.bkgd_ = {
        mode: VT100.A_NORMAL,
        fg: VT100.KY_FORE,
        bg: VT100.KY_BACK
    };
    this.c_attr_ = {
        mode: VT100.A_NORMAL,
        fg: VT100.KY_FORE,
        bg: VT100.KY_BACK
    };
    this.text_ = new Array(ht);
    this.attr_ = new Array(ht);
    for (r = 0; r < ht; ++r) {
        this.text_[r] = new Array(wd);
        this.attr_[r] = new Array(wd);
    }
    this.scr_ = scr;
    this.cursor_vis_ = true;
    this.cursor_col = 0;
    this.cursor_row = 0;
    this.grab_events_ = false;
    this.getch_isr_ = undefined;
    this.key_buf_ = [];
    this.echo_ = true;
    this.esc_state_ = 0;
    // Internal debug setting.
    this.debug_ = 0;
    this.clear();
    this.refresh();
}

// public constants -- colours and colour pairs
VT100.KY_BACK = 61;
VT100.KY_FORE = 62;

VT100.COLOR_BLACK = 0;
VT100.COLOR_BLUE = 1;
VT100.COLOR_GREEN = 2;
VT100.COLOR_CYAN = 3;
VT100.COLOR_RED = 4;
VT100.COLOR_MAGENTA = 5;
VT100.COLOR_YELLOW = 6;
VT100.COLOR_WHITE = 7;
VT100.COLOR_PAIRS = 256;
VT100.COLORS = 8;
// public constants -- attributes
VT100.A_NORMAL = 0;
VT100.A_UNDERLINE = 1;
VT100.A_REVERSE = 2;
VT100.A_BLINK = 4;
VT100.A_DIM = 8;
VT100.A_BOLD = 16;
VT100.A_STANDOUT = 32;
VT100.A_PROTECT = VT100.A_INVIS = 0; // ?
// other public constants
VT100.TABSIZE = 8;
// private constants
VT100.ATTR_FLAGS_ = VT100.A_UNDERLINE | VT100.A_REVERSE | VT100.A_BLINK |
    VT100.A_DIM | VT100.A_BOLD | VT100.A_STANDOUT |
    VT100.A_PROTECT | VT100.A_INVIS;
VT100.COLOR_SHIFT_ = 6;
// class variables
VT100.the_vt_ = undefined;

// class methods

// ***
// ***
// ***
// *** Translate Browser Events into vt100 terminal characters
// ***
VT100.InputString = function(e) {
    console.log("input " + e.key)
    if (e.ctrlKey) {
        console.log("ctrl " + e.key)
        let kk = (e.key).charCodeAt();
        if (e.ctrlKey && (kk >= 65 && kk <= 90)) {
            return String.fromCharCode(e.key - 64);
        }
        if (e.ctrlKey && (kk >= 97 && kk <= 122)) {
            return String.fromCharCode(kk - 96);
        }

        switch (e.key) {
            case $.Key2:
                return "\x00";
            case $.Key3:
            case $.LeftBracket:
                return "\x1b";
            case $.Key4:
                return "\x1c";
            case $.KeyRightBracket:
            case $.Key5:
                return "\x1d";
            case $.Key6:
                return "\x1e";
            case $.Key7:
                return "\x1f";
            case $.Key8:
                return "\x7f";
        }
    }

    console.log("Key");
    switch (e.key) {
        case "Backspace":
            return "\x08";
        case "Tab":
            return "\x09";
        case "Escape":
            return "\x1b";
        case "PageUp":
            return "\x1b[5~";
        case "PageDown":
            return "\x1b[6~";
        case "End":
            return "\x1b[4~";
        case "Home":
            return "\x1b[1~";
        case "LeftArrow", "ArrowLeft":
            return "\x1b[D";
        case "UpArrow", "ArrowUp":
            return "\x1b[A";
        case "RightArrow", "ArrowRight":
            return "\x1b[C";
        case "DownArrow", "ArrowDown":
            return "\x1b[B";
        case "Delete":
            return "\x7f";
        case "Enter":
            return "\x0a";
        case "Space":
            return " ";
            // case "F1:
            //     return "\x1b[11~";
            // case $.F2:
            //     return "\x1b[12~";
            // case $.F3:
            //     return "\x1b[13~";
            // case $.F4:
            //     return "\x1b[14~";
            // case $.F5:
            //     return "\x1b[15~";
            // case $.F6:
            //     return "\x1b[17~";
            // case $.F7:
            //     return "\x1b[18~";
            // case $.F8:
            //     return "\x1b[19~";
            // case $.F9:
            //     return "\x1b[20~";
            // case $.F10:
            //     return "\x1b[21~";
            // case $.F11:
            //     return "\x1b[23~";
            // case $.F12:
            //     return "\x1b[24~";
    }

    console.log("default " + e.key)
    return e.key; //e.chr;
}

// object methods

VT100.prototype.may_scroll_ = function() {
    var ht = this.ht_,
        cr = this.row_;
    while (cr >= ht) {
        this.scroll();
        --cr;
    }
    this.row_ = cr;
}

VT100.prototype.html_colours_ = function(attr) {
    var fg, bg, co0, co1;
    fg = attr.fg;
    bg = attr.bg;
    // is we are using Kris colors, and reverse, it's the cursor!
    if (((bg == VT100.KY_BACK) || (bg == VT100.KY_FORE)) && (attr.mode & VT100.A_REVERSE)) {
        return {
            f: '#AEAEAE',
            b: '#FFF926'
        };
    }
    // otherwise, its just a standard char
    if (bg == VT100.KY_BACK) {
        return {
            f: '#FFF926',
            b: '#AEAEAE'
        };
    }

    switch (attr.mode & (VT100.A_REVERSE | VT100.A_DIM | VT100.A_BOLD)) {
        case 0:
        case VT100.A_DIM | VT100.A_BOLD:
            co0 = '00';
            co1 = 'c0';
            break;
        case VT100.A_BOLD:
            co0 = '00';
            co1 = 'ff';
            break;
        case VT100.A_DIM:
            if (fg == VT100.COLOR_BLACK)
                co0 = '40';
            else
                co0 = '00';
            co1 = '40';
            break;
        case VT100.A_REVERSE:
        case VT100.A_REVERSE | VT100.A_DIM | VT100.A_BOLD:
            co0 = 'c0';
            co1 = '40';
            break;
        case VT100.A_REVERSE | VT100.A_BOLD:
            co0 = 'c0';
            co1 = '00';
            break;
        default:
            if (fg == VT100.COLOR_BLACK)
                co0 = '80';
            else
                co0 = 'c0';
            co1 = 'c0';
    }
    return {
        f: '#' + (fg & 4 ? co1 : co0) +
            (fg & 2 ? co1 : co0) +
            (fg & 1 ? co1 : co0),
        b: '        #' + (bg & 4 ? co1 : co0) +
            (bg & 2 ? co1 : co0) +
            (bg & 1 ? co1 : co0)
    };
}

VT100.prototype.addch = function(ch, attr) {
    var cc = this.col_;
    this.debug("addch:: ch: " + ch + ", attr: " + attr);
    switch (ch) {
        case '\b':
            if (cc != 0)
                --cc;
            break;
        case '\n':
            ++this.row_;
            cc = 0;
            this.clrtoeol();
            this.may_scroll_();
            break;
        case '\r':
            this.may_scroll_();
            cc = 0;
            break;
        case '\t':
            this.may_scroll_();
            cc += VT100.TABSIZE - cc % VT100.TABSIZE;
            if (cc >= this.wd_) {
                ++this.row_;
                cc -= this.wd_;
            }
            break;
        default:
            if (attr === undefined)
                attr = this._cloneAttr(this.c_attr_);
            if (cc >= this.wd_) {
                //console.log("3 incr this.row_");
                ++this.row_;
                cc = 0;
            }
            this.may_scroll_();
            this.text_[this.row_][cc] = ch;
            this.attr_[this.row_][cc] = attr;
            ++cc;
    }
    this.col_ = cc;
}

VT100.prototype.addstr = function(stuff) {
    for (var i = 0; i < stuff.length; ++i)
        this.addch(stuff.charAt(i));
}

VT100.prototype._cloneAttr = function VT100_cloneAttr(a) {
    return {
        mode: a.mode,
        fg: a.fg,
        bg: a.bg
    };
}

VT100.prototype.attroff = function(a) {
    //dump("attroff: " + a + "\n");
    a &= VT100.ATTR_FLAGS_;
    this.c_attr_.mode &= ~a;
}

VT100.prototype.attron = function(a) {
    //dump("attron: " + a + "\n");
    a &= VT100.ATTR_FLAGS_;
    this.c_attr_.mode |= a;
}

VT100.prototype.attrset = function(a) {
    //dump("attrset: " + a + "\n");
    this.c_attr_.mode = a;
}

VT100.prototype.fgset = function(fg) {
    //dump("fgset: " + fg + "\n");
    this.c_attr_.fg = fg;
}

VT100.prototype.bgset = function(bg) {
    //dump("bgset: " + bg + "\n");
    if (bg !== 0) {
        this.warn("bgset: " + bg + "\n");
    }
    this.c_attr_.bg = bg;
}

VT100.prototype.bkgdset = function(a) {
    this.bkgd_ = a;
}

VT100.prototype.clear = function() {
    this.debug("clear");
    this.row_ = this.col_ = 0;
    this.scrolled_ = 0;
    for (r = 0; r < this.ht_; ++r) {
        for (c = 0; c < this.wd_; ++c) {
            this.text_[r][c] = ' ';
            this.attr_[r][c] = this._cloneAttr(this.bkgd_);
        }
    }
}

VT100.prototype.clrtobot = function() {
    this.debug("clrtobot, row: " + this.row_);
    var ht = this.ht_;
    var wd = this.wd_;
    this.clrtoeol();
    for (var r = this.row_ + 1; r < ht; ++r) {
        for (var c = 0; c < wd; ++c) {
            this.text_[r][c] = ' ';
            this.attr_[r][c] = this.bkgd_;
        }
    }
}

VT100.prototype.clrtoeol = function() {
    this.debug("clrtoeol, col: " + this.col_);
    var r = this.row_;
    if (r >= this.ht_)
        return;
    for (var c = this.col_; c < this.wd_; ++c) {
        this.text_[r][c] = ' ';
        this.attr_[r][c] = this.bkgd_;
    }
}

VT100.prototype.clearpos = function(row, col) {
    this.debug("clearpos (" + row + ", " + col + ")");
    if (row < 0 || row >= this.ht_)
        return;
    if (col < 0 || col >= this.wd_)
        return;
    this.text_[row][col] = ' ';
    this.attr_[row][col] = this.bkgd_;
}

VT100.prototype.echo = function() {
    this.debug("echo on");
    this.echo_ = true;
}

VT100.prototype.erase = VT100.prototype.clear;

// VT100.prototype.getch = function(isr) {
//     this.debug("getch");
//     this.refresh();
//     this.getch_isr_ = isr;
//     setTimeout(VT100.go_getch_, 0);
// }

VT100.prototype.getmaxyx = function() {
    return { y: this.ht_ - 1, x: this.wd_ - 1 };
}

VT100.prototype.getyx = function() {
    return { y: this.row_, x: this.col_ };
}

VT100.prototype.move = function(r, c) {
    if (r < 0)
        r = 0;
    else if (r >= this.ht_)
        r = this.ht_ - 1;
    if (c < 0)
        c = 0;
    else if (c >= this.wd_)
        c = this.wd_ - 1;
    this.row_ = r;
    this.col_ = c;
    this.cursor_col = c;
    this.cursor_row = r;
    //console.log("move cursor: 2 (" + r + ", " + c + ")");
}

VT100.prototype.noecho = function() {
    this.debug("echo off");
    this.echo_ = false;
}

VT100.prototype.refresh = function() {
    this.debug("refresh");
    var r, c, stuff = "",
        start_tag = "",
        end_tag = "",
        at = -1,
        n_at, ch,
        pair, cr, cc, ht, wd, cv, added_end_tag;
    ht = this.ht_;
    wd = this.wd_;
    // cr = this.row_;
    // cc = this.col_;
    cr = this.cursor_row;
    cc = this.cursor_col;
    //console.log("(c,r) ", cc, cr);
    cv = this.cursor_vis_;
    var innerHTML = this.scr_.innerHTML;
    if (cc >= wd)
        cc = wd - 1;
    for (r = 0; r < ht; ++r) {
        if (r > 0) {
            stuff += '\n';
        }
        for (c = 0; c < wd; ++c) {
            added_end_tag = false;
            n_at = this.attr_[r][c];
            if (cv && r == cr && c == cc) {
                // Draw the cursor here.
                n_at = this._cloneAttr(n_at);
                n_at.mode ^= VT100.A_REVERSE;
                console.log("cursor " + c + ":" + r)
            }
            // If the attributes changed, make a new span.
            if (n_at.mode != at.mode || n_at.fg != at.fg || n_at.bg != at.bg) {
                if (c > 0) {
                    stuff += end_tag;
                }
                start_tag = "";
                end_tag = "";
                // if (n_at.mode & VT100.A_BLINK) {
                //     start_tag = "<blink>"; // blink tags considered harmful.
                //     end_tag = "</blink>" + end_tag; // and deprecated as well.
                // }
                if (n_at.mode & VT100.A_STANDOUT)
                    n_at.mode |= VT100.A_BOLD;
                pair = this.html_colours_(n_at);
                start_tag += '<span style="color:' + pair.f +
                    ';background-color:' + pair.b;
                if (n_at.mode & VT100.A_UNDERLINE)
                    start_tag += ';text-decoration:underline';
                start_tag += ';">';
                stuff += start_tag;
                end_tag = "</span>" + end_tag;
                at = n_at;
                added_end_tag = true;
            } else
            if (c == 0) {
                stuff += start_tag;
            }
            ch = this.text_[r][c];
            switch (ch) {
                case '&':
                    stuff += '&amp;';
                    break;
                case '<':
                    stuff += '&lt;';
                    break;
                case '>':
                    stuff += '&gt;';
                    break;
                case ' ':
                    //stuff += '&nbsp;';	break;
                    stuff += ' ';
                    break;
                default:
                    stuff += ch;
            }
        }
        if (!added_end_tag)
            stuff += end_tag;
    }
    this.scr_.innerHTML = "<b>" + stuff + "</b>\n";
}

VT100.prototype.scroll = function() {
    this.scrolled_ += 1;
    this.debug("scrolled: " + this.scrolled_);
    var n_text = this.text_[0],
        n_attr = this.attr_[0],
        ht = this.ht_,
        wd = this.wd_;
    for (var r = 1; r < ht; ++r) {
        this.text_[r - 1] = this.text_[r];
        this.attr_[r - 1] = this.attr_[r];
    }
    this.text_[ht - 1] = n_text;
    this.attr_[ht - 1] = n_attr;
    for (var c = 0; c < wd; ++c) {
        n_text[c] = ' ';
        n_attr[c] = this.bkgd_;
    }
}

VT100.prototype.standend = function() {
    //this.debug("standend");
    this.attrset(0);
}

VT100.prototype.standout = function() {
    //this.debug("standout");
    this.attron(VT100.A_STANDOUT);
}

VT100.prototype.write = function(stuff) {
    console.log(">> " + stuff)
    var ch, x, r, c, i, j, yx, myx;
    for (i = 0; i < stuff.length; ++i) {
        ch = stuff.charAt(i);
        if (ch == '\x0D') {
            this.debug("write:: ch: " + ch.charCodeAt(0) + ", '\\x0D'");
        } else {
            this.debug("write:: ch: " + ch.charCodeAt(0) + ", '" + (ch == '\x1b' ? "ESC" : ch) + "'");
        }
        //dump("ch: " + ch.charCodeAt(0) + ", '" + (ch == '\x1b' ? "ESC" : ch) + "'\n");
        switch (ch) {
            case '\x00':
            case '\x7f':
            case '\x07':
                /* bell, ignore it */
                this.debug("write:: ignoring bell character: " + ch);
                continue;
            case '\a':
            case '\b':
            case '\t':
            case '\r':
                this.addch(ch);
                continue;
            case '\n':
            case '\v':
            case '\f': // what a mess
                yx = this.getyx();
                myx = this.getmaxyx();
                if (yx.y >= myx.y) {
                    this.scroll();
                    this.move(myx.y, 0);
                } else
                    this.move(yx.y + 1, 0);
                continue;
            case '\x18':
            case '\x1a':
                this.esc_state_ = 0;
                this.debug("write:: set escape state: 0");
                continue;
            case '\x1b':
                this.esc_state_ = 1;
                this.debug("write:: set escape state: 1");
                continue;
            case '\x9b':
                this.esc_state_ = 2;
                this.debug("write:: set escape state: 2");
                continue;
        }
        // not a recognized control character
        switch (this.esc_state_) {
            case 0: // not in escape sequence
                this.addch(ch);
                break;
            case 1: // just saw ESC
                switch (ch) {
                    case '[':
                        this.esc_state_ = 2;
                        this.debug("write:: set escape state: 2");
                        break;
                    case '=':
                        /* Set keypade mode (ignored) */
                        this.debug("write:: set keypade mode: ignored");
                        this.esc_state_ = 0;
                        break;
                    case '>':
                        /* Reset keypade mode (ignored) */
                        this.debug("write:: reset keypade mode: ignored");
                        this.esc_state_ = 0;
                        break;
                    case 'H':
                        /* Set tab at cursor column (ignored) */
                        this.debug("write:: set tab cursor column: ignored");
                        this.esc_state_ = 0;
                        break;
                }
                break;
            case 2: // just saw CSI
                switch (ch) {
                    case 'K':
                        /* Erase in Line */
                        this.esc_state_ = 0;
                        this.clrtoeol();
                        continue;
                    case 'H':
                        /* Move to (0,0). */
                        this.esc_state_ = 0;
                        //console.log("set cursor home");
                        this.move(0, 0);
                        continue;
                    case 'J':
                        /* Clear to the bottom. */
                        this.esc_state_ = 0;
                        this.clrtobot();
                        continue;
                    case '?':
                        /* Special VT100 mode handling. */
                        this.esc_state_ = 5;
                        this.debug("write:: special vt100 mode");
                        continue;
                }
                // Drop through to next case.
                this.csi_parms_ = [0];
                this.debug("write:: set escape state: 3");
                this.esc_state_ = 3;
            case 3: // saw CSI and parameters
                switch (ch) {
                    case '0':
                    case '1':
                    case '2':
                    case '3':
                    case '4':
                    case '5':
                    case '6':
                    case '7':
                    case '8':
                    case '9':
                        x = this.csi_parms_.pop();
                        this.csi_parms_.push(x * 10 + ch * 1);
                        this.debug("csi_parms_: " + this.csi_parms_);
                        continue;
                    case ';':
                        if (this.csi_parms_.length < 17)
                            this.csi_parms_.push(0);
                        continue;
                }
                this.esc_state_ = 0;
                switch (ch) {
                    case 'A':
                        // Cursor Up 		<ESC>[{COUNT}A
                        this.move(this.row_ - Math.max(1, this.csi_parms_[0]),
                            this.col_);
                        break;
                    case 'B':
                        // Cursor Down 		<ESC>[{COUNT}B
                        this.move(this.row_ + Math.max(1, this.csi_parms_[0]),
                            this.col_);
                        break;
                    case 'C':
                        // Cursor Forward 	<ESC>[{COUNT}C
                        this.move(this.row_,
                            this.col_ + Math.max(1, this.csi_parms_[0]));
                        break;
                    case 'c':
                        this.warn("write:: got TERM query");
                        break;
                    case 'D':
                        // Cursor Backward 	<ESC>[{COUNT}D
                        this.move(this.row_,
                            this.col_ - Math.max(1, this.csi_parms_[0]));
                        break;
                    case 'f':
                    case 'H':
                        // Cursor Home 		<ESC>[{ROW};{COLUMN}H
                        //console.log("set cursor position");
                        this.csi_parms_.push(0);
                        this.move(this.csi_parms_[0],
                            this.csi_parms_[1]);
                        break;
                    case 'J':
                        switch (this.csi_parms_[0]) {
                            case 0:
                                this.clrtobot();
                                break;
                            case 2:
                                this.clear();
                                this.move(0, 0);
                        }
                        break;
                    case 'm':
                        for (j = 0; j < this.csi_parms_.length; ++j) {
                            x = this.csi_parms_[j];
                            switch (x) {
                                case 0:
                                    this.standend();
                                    this.fgset(this.bkgd_.fg);
                                    this.bgset(this.bkgd_.bg);
                                    break;
                                case 1:
                                    this.attron(VT100.A_BOLD);
                                    break;
                                case 30:
                                    this.fgset(VT100.COLOR_BLACK);
                                    break;
                                case 31:
                                    this.fgset(VT100.COLOR_RED);
                                    break;
                                case 32:
                                    this.fgset(VT100.COLOR_GREEN);
                                    break;
                                case 33:
                                    this.fgset(VT100.COLOR_YELLOW);
                                    break;
                                case 34:
                                    this.fgset(VT100.COLOR_BLUE);
                                    break;
                                case 35:
                                    this.fgset(VT100.COLOR_MAGENTA);
                                    break;
                                case 36:
                                    this.fgset(VT100.COLOR_CYAN);
                                    break;
                                case 37:
                                    this.fgset(VT100.COLOR_WHITE);
                                    break;
                                case 40:
                                    this.bgset(VT100.COLOR_BLACK);
                                    break;
                                case 41:
                                    this.bgset(VT100.COLOR_RED);
                                    break;
                                case 42:
                                    this.bgset(VT100.COLOR_GREEN);
                                    break;
                                case 44:
                                    this.bgset(VT100.COLOR_YELLOW);
                                    break;
                                case 44:
                                    this.bgset(VT100.COLOR_BLUE);
                                    break;
                                case 45:
                                    this.bgset(VT100.COLOR_MAGENTA);
                                    break;
                                case 46:
                                    this.bgset(VT100.COLOR_CYAN);
                                    break;
                                case 47:
                                    this.bgset(VT100.COLOR_WHITE);
                                    break;
                            }
                        }
                        break;
                    case 'r':
                        // 1,24r - set scrolling region (ignored)
                        break;
                    case '[':
                        this.debug("write:: set escape state: 4");
                        this.esc_state_ = 4;
                        break;
                    case 'g':
                        // 0g: clear tab at cursor (ignored)
                        // 3g: clear all tabs (ignored)
                        break;
                    default:
                        this.warn("write:: unknown command: 0x" + ch.charCodeAt(0).toString(16));
                        this.csi_parms_ = [];
                        break;
                }
                break;
            case 4: // saw CSI [
                this.esc_state_ = 0; // gobble char.
                break;
            case 5: // Special mode handling, saw <ESC>[?
                // Expect a number - the reset type
                this.csi_parms_ = [ch];
                this.esc_state_ = 6;
                break;
            case 6: // Reset mode handling, saw <ESC>[?1
                // Expect a letter - the mode target, example:
                // <ESC>[?1l : cursor key mode = cursor
                // <ESC>[?1h : save current screen, create new empty
                //             screen and position at 0,0
                // <ESC>[?5l : White on blk
                // XXX: Ignored for now.
                //dump("Saw reset mode: <ESC>[?" + this.csi_parms_[0] + ch + "\n");
                this.esc_state_ = 0;
                this.debug("write:: set escape state: 0");
                break;
        }
    }
    this.refresh();
}

VT100.prototype.debug = function(message) {
    if (this.debug_) {
        console.log(message + "\n");
    }
}

VT100.prototype.warn = function(message) {
    console.log(message + "\n");
}