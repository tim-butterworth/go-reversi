package main

import (
	"fmt"
	"net/http"
	"reversi/tcpimpl"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Text string
}

type WebSocketPlayer struct {
	connection *websocket.Conn
}

func (player *WebSocketPlayer) Error(message string) {
	player.connection.WriteJSON(Message{Text: message})
}

func (player *WebSocketPlayer) Message(message string) {
	player.connection.WriteJSON(Message{Text: message})
}

func (player *WebSocketPlayer) ReadJson(container interface{}) error {
	return player.connection.ReadJSON(&container)
}

func (player *WebSocketPlayer) Close() {
	player.connection.Close()
}

func main() {
	fmt.Println("Should start up a websocket server!")

	fs := http.FileServer(http.Dir("../../jsApp"))

	http.Handle("/app/", http.StripPrefix("/app", fs))
	http.HandleFunc("/ws", getHandleConnections())

	host := ":9090"
	fmt.Printf("http server started on %s", host)
	err := http.ListenAndServe(host, nil)
	if err != nil {
		fmt.Println("ListenAndServe: ", err)
	}
}

func getHandleConnections() func(http.ResponseWriter, *http.Request) {
	pendingGame := tcpimpl.NewPendingGame()

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("WS connection requested")
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err.Error())
		}
		webSocketPlayer := WebSocketPlayer{
			connection: ws,
		}
		pendingGame.AddPlayer(&webSocketPlayer)

		if pendingGame.IsFull() {
			activeGame := tcpimpl.NewActiveGame(pendingGame)

			go activeGame.Start()
		}
	}
}
