package buffer

import "testing"

func TestLoadFile(t *testing.T) {
	fname := "testtext.txt"

	tt := LoadFile(fname)
	if tt.size() <= 0 {
		t.Errorf("File Size is bad")
	}

}
