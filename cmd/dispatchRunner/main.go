package main

import (
	"fmt"
	"reversi/core/dispatch"
	"time"
)

type Actor struct {
	id            string
	acceptChannel chan dispatch.Message
}

func NewActor(name string) Actor {
	return Actor{
		id:            name,
		acceptChannel: make(chan dispatch.Message),
	}
}

func runActor(actor Actor, messageFun func(dispatch.ActorMessage)) {
	for {
		select {
		case message := <-actor.acceptChannel:
			fmt.Println(message.Source)
			fmt.Println(message.Body)

			messageFun(dispatch.ActorMessage{
				Body:        "Hi there!",
				Destination: message.Source,
			})
		}
	}
}

func wait(seconds int32) {
	time.Since(time.Now().Add(time.Second * 10))
	<-time.After(time.Second * time.Duration(seconds))
}

func main() {
	dispatcher := dispatch.NewDispatch()
	dispatcher.Start()

	actorCount := 20
	actors := make([]Actor, actorCount)
	actorMessageSenders := make([]func(dispatch.ActorMessage), actorCount)

	for i := range actors {
		actors[i] = NewActor(fmt.Sprintf("%d", i))

		messageFun := dispatcher.Register(actors[i].acceptChannel, actors[i].id)
		actorMessageSenders[i] = messageFun

		go runActor(actors[i], messageFun)
	}

	fmt.Println("Got here")

	actorMessageSenders[0](dispatch.ActorMessage{
		Destination: "10",
		Body:        "This is from 10...",
	})

	wait(10)
}
