package web

import (
	"log"
	"net/http"
)

func echoserver() {
	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity

		if err != nil {
			panic("Nope. ")
		}
		log.Println("going into echo loop")
		for {
			// Read message from browser
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			// Print the message to the console
			log.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))

			// Write message back to browser
			if err = conn.WriteMessage(msgType, msg); err != nil {
				return
			}
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "websocket.html")
	})
	http.HandleFunc("/quit", func(w http.ResponseWriter, r *http.Request) {
		panic("quitting websocket")
	})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web"))))

	http.ListenAndServe(":8005", nil)
}
