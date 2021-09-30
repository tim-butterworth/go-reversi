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

type wrapper struct {
	wrapped http.Handler
}

func (self *wrapper) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	fmt.Println(request.RequestURI)

	self.wrapped.ServeHTTP(writer, request)
}

func getWrapper(wrapped http.Handler) http.Handler {
	return &wrapper{
		wrapped: wrapped,
	}
}

type PlayerMessage struct {
	playerId string
	body     interface{}
}

func MakePlayerMessage(playerId string, body interface{}) PlayerMessage {
	return PlayerMessage{
		playerId: playerId,
		body:     body,
	}
}

type PlayerConnection struct {
	playerId   string
	connection *websocket.Conn
}

type IDProvider interface {
	getId() string
}

type incrementIdProvider struct {
	currentValue int64
}

func (self *incrementIdProvider) getId() string {
	id := fmt.Sprintf("%d", self.currentValue)
	self.currentValue += 1

	return id
}
func NewIncrementIdProvider() IDProvider {
	return &incrementIdProvider{}
}

func handleConnectionRegistration(connectionChannels ConnectionChannels, idProvider IDProvider) {
	for {
		select {
		case ws := <-connectionChannels.newConnections:
			connectionChannels.playerConnection <- PlayerConnection{
				playerId:   idProvider.getId(),
				connection: ws,
			}
		case connection := <-connectionChannels.reconnections:
			connectionChannels.playerConnection <- connection
		}
	}
}

func playerOutput(outgoingMessages <-chan interface{}, removePlayer chan<- string, playerConnection PlayerConnection) {
	socket := playerConnection.connection
	socket.SetCloseHandler(func(code int, text string) error {
		removePlayer <- playerConnection.playerId

		return nil
	})

	for {
		outgoingMessage := <-outgoingMessages

		err := socket.WriteJSON(outgoingMessage)
		if err != nil {
			fmt.Println(fmt.Sprintf("Error writing message to websocket for player id [%s]", playerConnection.playerId), err)
		}
	}
}

func readFromPlayerSocket(inChannel chan<- string, connection *websocket.Conn) {
	for {
		connection.ReadJSON(&Message{})
	}
}

func playerInput(terminateChannel <-chan bool, playerConnection PlayerConnection) {
	defer playerConnection.connection.Close()

	inChannel := make(chan string)
	go readFromPlayerSocket(inChannel, playerConnection.connection)

	for {
		select {
		case <-terminateChannel:
			break
		case str := <-inChannel:
			fmt.Println(str)
		}
	}

	fmt.Printf("Terminating input go-routine for player [%s]\n", playerConnection.playerId)
}

type OutBoundMessage struct {
	PlayerId string
	Message  interface{}
}

type Command struct {
	CommandType string
	Body        []byte
}
type InBoundMessage struct {
	PlayerId string
	Command  Command
}

type PlayerChannels struct {
	Outbound  chan<- interface{}
	Terminate chan<- bool
}

func handleConnections(playerConnections <-chan PlayerConnection, outboundMessages <-chan OutBoundMessage, inBoundMessageChannel chan<- InBoundMessage) {
	playerConnectionsById := make(map[string]PlayerChannels)
	removePlayer := make(chan string)

	for {
		select {
		case playerConnection := <-playerConnections:

			playerOutChannel := make(chan interface{})
			playerTerminateChannel := make(chan bool)

			playerConnectionsById[playerConnection.playerId] = PlayerChannels{
				Outbound:  playerOutChannel,
				Terminate: playerTerminateChannel,
			}

			go playerOutput(playerOutChannel, removePlayer, playerConnection)
			go playerInput(playerTerminateChannel, playerConnection)
		case idToRemove := <-removePlayer:
			channels, exists := playerConnectionsById[idToRemove]
			if exists {
				delete(playerConnectionsById, idToRemove)
				channels.Terminate <- true
			}
		case outboundMessage := <-outboundMessages:
			playerId := outboundMessage.PlayerId
			message := outboundMessage.Message

			channels, exists := playerConnectionsById[playerId]
			if exists {
				select {
				case channels.Outbound <- message:
				default:
					fmt.Println("No message was sent")
				}
			}
		}
	}
}

type ConnectionChannels struct {
	newConnections   <-chan *websocket.Conn
	reconnections    <-chan PlayerConnection
	playerConnection chan<- PlayerConnection
}

func main() {
	fmt.Println("Should start up a websocket server!")

	fs := http.FileServer(http.Dir("../../jsApp"))

	newConnectionsChannel := make(chan *websocket.Conn)
	reconnectionsChannel := make(chan PlayerConnection)
	playerConnectionChannel := make(chan PlayerConnection)

	connectionChannels := ConnectionChannels{
		newConnections:   newConnectionsChannel,
		reconnections:    reconnectionsChannel,
		playerConnection: playerConnectionChannel,
	}

	outboundMessageChannel := make(chan OutBoundMessage)
	inboundMessageChannel := make(chan InBoundMessage)

	go handleConnectionRegistration(connectionChannels, NewIncrementIdProvider())
	go handleConnections(playerConnectionChannel, outboundMessageChannel, inboundMessageChannel)

	http.Handle("/app/", http.StripPrefix("/app", getWrapper(fs)))
	http.HandleFunc("/ws", getHandleConnections())
	http.HandleFunc("/game/create/", getHandleCreateGame())
	http.HandleFunc("/ows", getHandleOtherConnections(newConnectionsChannel, reconnectionsChannel))

	host := ":9090"
	fmt.Printf("http server started on %s\n", host)
	err := http.ListenAndServe(host, nil)
	if err != nil {
		fmt.Println("ListenAndServe: ", err)
	}
}

func getHandleCreateGame() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("New game requested")

		w.Header().Add("status", "200")
		w.Write([]byte("success"))
	}
}

func getHandleOtherConnections(newConnectionChannel chan<- *websocket.Conn, reconnectionChannel chan<- PlayerConnection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err.Error())
			w.Header().Add("status", "500")
			w.Write([]byte("Error"))
			return
		}

		newConnectionChannel <- ws
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
