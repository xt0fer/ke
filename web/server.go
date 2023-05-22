package web

import (
	"log"
	"net/http"

	"github.com/kristofer/ke/editor"
	"github.com/kristofer/ke/term"

	"github.com/gorilla/websocket"
)

// We'll need to define an Upgrader
// this will require a Read and Write buffer size
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func echoserver() {
	http.HandleFunc("/editor", func(w http.ResponseWriter, r *http.Request) {
		log.Println("serving editor page")

		conn, err := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity

		if err != nil {
			panic("Nope. ")
		}
		log.Println("going into editor loop")
		editor := editor.NewEditor()
		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			event := editor.Term.EventFromKey(msg)

			ok := editor.HandleEvent(event)
			if !ok {
				return //exit editor
			}

			msg = []byte(term.CUP(0, 0))
			msg = append(msg, []byte(term.ED(term.EraseToEnd))...)
			s := editor.RootBuffer.T.AllContents()
			msg = append(msg, []byte(s)...)
			//log.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))

			if err = conn.WriteMessage(msgType, msg); err != nil {
				return
			}
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("serving root page")

		http.ServeFile(w, r, "static/editor.html")
	})
	http.HandleFunc("/quit", func(w http.ResponseWriter, r *http.Request) {
		panic("quitting websocket")
	})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.ListenAndServe(":8005", nil)
}
