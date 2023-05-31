package web

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kristofer/ke/kg"

	"github.com/gorilla/websocket"
)

// We'll need to define an Upgrader
// this will require a Read and Write buffer size
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (editor *EditorServer) kgEditor(w http.ResponseWriter, r *http.Request) {

	log.Println("running KG editor")

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("Nope. No websocket created. see editorserver()")
		return
	}

	go func() {
		editor.Editor = &kg.Editor{}
		editor.Editor.StartEditor([]string{}, 0, conn, editor.Quit)

		//editor.Quit <- syscall.SIGINT
	}()

	log.Println("ending KG editor")

}

type EditorServer struct {
	Server *http.Server
	Editor *kg.Editor
	Quit   chan os.Signal
}

func NewEditorServer() *EditorServer {
	e := &EditorServer{}
	e.Server = &http.Server{
		Addr: ":8005",
	}
	e.Quit = make(chan os.Signal, 1)
	return e
}

func (editor *EditorServer) StartEditorServer() {

	http.HandleFunc("/editor", editor.kgEditor)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("serving main page")

		http.ServeFile(w, r, "static/vt100.html")
	})

	http.HandleFunc("/vt100", func(w http.ResponseWriter, r *http.Request) {
		log.Println("serving main page")

		http.ServeFile(w, r, "static/vt100.html")
	})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	//http.ListenAndServe(":8005", nil)
	go func() {
		if err := editor.Server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
		log.Println("Stopped serving new connections.")
	}()
	//	sigChan := make(chan os.Signal, 1)
	signal.Notify(editor.Quit, syscall.SIGINT, syscall.SIGTERM)
	<-editor.Quit

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := editor.Server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}
	log.Println("Graceful shutdown complete.")

}

// func keEditor(w http.ResponseWriter, r *http.Request) {

// 	conn, err := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity

// 	if err != nil {
// 		log.Println("Nope. No websocket created. see editorserver()")
// 		return
// 	}
// 	//log.Println("going into editor loop")
// 	editor := editor.NewEditor()

// 	m := editor.DisplayContents(editor.CurrentScreen())

// 	if err = conn.WriteMessage(1, m); err != nil {
// 		log.Println("writing new editor failed.")
// 		return
// 	}

// 	for { // event loop, server side
// 		msgType, msg, err := conn.ReadMessage()
// 		if err != nil {
// 			log.Println("unable to get message from frontend")
// 			return
// 		}

// 		event := editor.Term.EventFromKey(msg)

// 		ok := editor.HandleEvent(event)
// 		if !ok {
// 			msg := editor.DisplayContents("Exiting...")
// 			if err = conn.WriteMessage(msgType, msg); err != nil {
// 				log.Println("unable to write [Exiting...]")
// 			}
// 			conn.Close()
// 			break //exit editor
// 		}

// 		msg = editor.DisplayContents(editor.CurrentScreen())

// 		if err = conn.WriteMessage(msgType, msg); err != nil {
// 			log.Println("unable to write message to frontend")
// 			return
// 		}
// 	}
// }
