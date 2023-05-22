package web

import (
	"log"
	"net/http"

	"github.com/kristofer/ke/editor"

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
		//log.Println("going into editor loop")
		editor := editor.NewEditor()

		m := editor.DisplayContents(editor.CurrentScreen())

		if err = conn.WriteMessage(1, m); err != nil {
			log.Println("writing new editor failed.")
			return
		}

		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("unable to get message from frontend")
				return
			}

			event := editor.Term.EventFromKey(msg)

			ok := editor.HandleEvent(event)
			if !ok {
				msg := editor.DisplayContents("Exiting...")
				if err = conn.WriteMessage(msgType, msg); err != nil {
				}
				conn.Close()
				break //exit editor
			}

			msg = editor.DisplayContents(editor.CurrentScreen())

			if err = conn.WriteMessage(msgType, msg); err != nil {
				log.Println("unable to write message to frontend")
				return
			}
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("starting up frontend")

		http.ServeFile(w, r, "static/editor.html")
	})
	http.HandleFunc("/quit", func(w http.ResponseWriter, r *http.Request) {
		panic("quitting websocket")
	})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.ListenAndServe(":8005", nil)
}
