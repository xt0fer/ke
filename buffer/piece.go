package buffer

func (p *Piece) size() int {
	return p.Run
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
	left = NewPiece(p.Source, p.Start, idx)
	right = NewPiece(p.Source, p.Start+idx, p.Run-idx)
	return
}
