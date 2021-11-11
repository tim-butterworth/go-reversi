package dispatch

import "fmt"

type Message struct {
	Source string
	Body   interface{}
}

type ActorMessage struct {
	Destination string
	Body        interface{}
}

type MessageContainer struct {
	Destination string
	Message     Message
}

type Dispatch interface {
	Register(accept chan<- Message, id string) func(ActorMessage)
	Start()
}

type Messageable struct {
	id     string
	accept chan<- Message
}

type dispatch struct {
	publish      chan MessageContainer
	actors       map[string]chan<- Message
	registration chan Messageable
}

func (self *dispatch) Register(accept chan<- Message, id string) func(ActorMessage) {
	self.actors[id] = accept

	self.registration <- Messageable{
		id:     id,
		accept: accept,
	}

	return func(message ActorMessage) {
		self.publish <- MessageContainer{
			Destination: message.Destination,
			Message: Message{
				Source: id,
				Body:   message.Body,
			},
		}
	}
}

func run(dispatch *dispatch) {
	fmt.Println("Dispatch started")
	for {
		select {
		case messageable := <-dispatch.registration:
			dispatch.actors[messageable.id] = messageable.accept
		case message := <-dispatch.publish:
			destination := message.Destination
			receiver, exists := dispatch.actors[destination]

			if exists {
				receiver <- message.Message
			} else {
				fmt.Println("That id does not exist")
			}
		}
	}
}

func (self *dispatch) Start() {
	go run(self)
}

func NewDispatch() Dispatch {
	return &dispatch{
		publish:      make(chan MessageContainer),
		actors:       make(map[string]chan<- Message),
		registration: make(chan Messageable),
	}
}
