package web

import (
	"log"
	"net/http"

	"github.com/kristofer/ke/editor"
	"github.com/kristofer/ke/kg"

	"github.com/gorilla/websocket"
)

// We'll need to define an Upgrader
// this will require a Read and Write buffer size
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func kgEditor(w http.ResponseWriter, r *http.Request) {
	//edit.StartEditor([]string{}, 0)

	log.Println("running KG editor")

	conn, err := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity

	if err != nil {
		log.Println("Nope. No websocket created. see editorserver()")
		return
	}
	//log.Println("going into editor loop")
	go func() {
		edit := &kg.Editor{}
		edit.StartEditor([]string{}, 0, conn)
	}()

}

func keEditor(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity

	if err != nil {
		log.Println("Nope. No websocket created. see editorserver()")
		return
	}
	//log.Println("going into editor loop")
	editor := editor.NewEditor()

	m := editor.DisplayContents(editor.CurrentScreen())

	if err = conn.WriteMessage(1, m); err != nil {
		log.Println("writing new editor failed.")
		return
	}

	for { // event loop, server side
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
				log.Println("unable to write [Exiting...]")
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
}

func EditorServer() {
	http.HandleFunc("/editor", kgEditor)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("serving main page")

		http.ServeFile(w, r, "static/editor.html")
	})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.ListenAndServe(":8005", nil)
}
