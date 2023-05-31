package web

import "testing"

func TestVT100server(t *testing.T) {

	es := NewEditorServer()
	es.StartEditorServer()
}
