package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type webMsg struct {
	Date     string
	From     string
	Subject  string
	Message  string
	Original string
}

type webAPI struct {
	listeners []chan webMsg
}

func newWebAPI() *webAPI {
	web := &webAPI{
		listeners: []chan webMsg{},
	}

	go func() {
		if err := web.serve(); err != nil {
			log.Fatalf("Could not run web: %v\n", err)
		}
	}()
	return web
}

func (web *webAPI) serve() error {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	fs := http.FileServer(http.Dir("./html"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Websocket Connected!")
		web.listen(websocket)
	})
	return http.ListenAndServe(":8080", nil)
}

func (web *webAPI) push(from, date, subject, message string, original string) error {
	log.Printf("Listeners: %d\n", len(web.listeners))
	for i := range web.listeners {
		msg := webMsg{
			Date:     date,
			From:     from,
			Subject:  subject,
			Message:  message,
			Original: original,
		}
		select {
		case web.listeners[i] <- msg:
		default:
		}
	}
	return nil
}

func (web *webAPI) listen(conn *websocket.Conn) {
	outCh := make(chan webMsg, 1)
	web.listeners = append(web.listeners, outCh)
	for msg := range outCh {
		b, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Error marshaling message: %v\n", err)
			continue
		}
		if err := conn.WriteMessage(1, b); err != nil {
			log.Printf("Client disconnected: %v\n", err)
			return
		}
	}
}
