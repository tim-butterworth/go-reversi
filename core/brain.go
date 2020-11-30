package core

type GameBrain interface {
	Initialize(rejectHandler CommandRejectHandler)
	ExecuteCommand(move Command, rejectHandler CommandRejectHandler)
}

type gameBrain struct {
	commandHandler *CommandHandler
}

type multipleConsumerAdapter struct {
	consumers []EventConsumer
}

func (multiConsumer *multipleConsumerAdapter) SendEvent(event Event) {
	for _, consumer := range multiConsumer.consumers {
		consumer.SendEvent(event)
	}
}

func (brain *gameBrain) Initialize(rejectHandler CommandRejectHandler) {
	brain.commandHandler.AttemptCommand(NewInitializeCommand(), rejectHandler)
}

func (brain *gameBrain) ExecuteCommand(move Command, rejectHandler CommandRejectHandler) {
	brain.commandHandler.AttemptCommand(move, rejectHandler)
}

func NewGameBrain(eventConsumer EventConsumer) GameBrain {
	gameEventAggregator := NewGameEventAggregator()
	commandHandler := NewCommandHandler(
		&multipleConsumerAdapter{
			consumers: []EventConsumer{gameEventAggregator, eventConsumer},
		},
		gameEventAggregator,
	)

	gameEventAggregator.Register(NewPrintStateConsumer())

	return &gameBrain{commandHandler: commandHandler}
}
