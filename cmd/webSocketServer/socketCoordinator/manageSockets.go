package socketcoordinator

import (
	"encoding/json"
	"fmt"
)

type InboundMessage struct {
	path []string
	body json.RawMessage
}

type OutboundMessage struct {
	recipientId string
	message     interface{}
}

type IdentifiedInboundMessage struct {
	Id      string
	Message InboundMessage
}
type MessageSource interface {
	pushMessage(InboundMessage)
}

type SocketWrapper interface {
	Id() string
	Write(interface{}) error
	setSource(MessageSource)
}

type JoinArguments struct {
	JoinableId string
	SocketId   string
}

type Joinable interface {
	Id() string
	joinIfMatches(JoinArguments)
	setOutboundChannel(chan<- OutboundMessage)
	applyCommand(string, interface{})
}

type SocketManager interface {
	RegisterSocket(SocketWrapper)
	RegisterJoinable(Joinable)
	Run()
}

type socketManager struct {
	joinableMapping  map[string]Joinable
	socketMapping    map[string]SocketWrapper
	outboundMessages chan OutboundMessage
	inboundMessages  chan IdentifiedInboundMessage
}

type messageSource struct {
	inChannel chan<- IdentifiedInboundMessage
	id        string
}

func (self messageSource) pushMessage(rawMessage InboundMessage) {
	self.inChannel <- IdentifiedInboundMessage{
		Id:      self.id,
		Message: rawMessage,
	}
}

func newMessateSource(id string, inChannel chan<- IdentifiedInboundMessage) MessageSource {
	return &messageSource{
		id:        id,
		inChannel: inChannel,
	}
}

func (self *socketManager) RegisterSocket(socket SocketWrapper) {
	self.socketMapping[socket.Id()] = socket
	socket.setSource(newMessateSource(socket.Id(), self.inboundMessages))
}

func (self *socketManager) RegisterJoinable(joinable Joinable) {
	self.joinableMapping[joinable.Id()] = joinable
	joinable.setOutboundChannel(self.outboundMessages)
}

func withSocket(manager *socketManager, socketId string, ifFound func(SocketWrapper) error) error {
	if socket, found := manager.socketMapping[socketId]; found {
		return ifFound(socket)
	}

	fmt.Println(fmt.Sprintf("Failed to find socket with id: [%s]", socketId))
	return nil
}

func handleInboundMessages(manager *socketManager) {
	for {
		message := <-manager.inboundMessages

		socketId := message.Id
		path := message.Message.path
		body := message.Message.body
		pathLength := len(path)

		if pathLength == 0 {
			// should be a join command or maybe a list all joinables command
		} else if pathLength == 1 {
			joinableId := path[0]
			if joinable, found := manager.joinableMapping[joinableId]; found {
				joinable.applyCommand(socketId, body)
			} else {
				err := withSocket(manager, socketId, func(socket SocketWrapper) error {
					return socket.Write("Some error message about that particular id not existing")
				})
				if err != nil {
					fmt.Println("Error", err)
				}
			}
		} else {
			err := withSocket(manager, socketId, func(socket SocketWrapper) error {
				return socket.Write("Invalid command, path should be empty or contain the id of a joinable")
			})
			if err != nil {
				fmt.Println("Error", err)
			}
		}
	}
}

func handleOutboundMessages(manager *socketManager) {
	for {
		message := <-manager.outboundMessages
		socketId := message.recipientId

		err := withSocket(manager, socketId, func(socket SocketWrapper) error {
			return socket.Write(message.message)
		})

		if err != nil {
			// maybe this means the socket should be closed and removed from the map
			fmt.Println("Error", err)
		}
	}
}

func (self *socketManager) Run() {
	go handleInboundMessages(self)
	go handleOutboundMessages(self)
}

func NewSocketManager() SocketManager {
	return &socketManager{
		joinableMapping:  make(map[string]Joinable),
		socketMapping:    make(map[string]SocketWrapper),
		outboundMessages: make(chan OutboundMessage),
		inboundMessages:  make(chan IdentifiedInboundMessage),
	}
}
