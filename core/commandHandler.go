package core

type CommandType string

const (
	INITIALIZE CommandType = "INITIALIZE"
	MOVE       CommandType = "MOVE"
	CONCEDE    CommandType = "CONCEDE"
)

type Command struct {
	commandType CommandType
	data        interface{}
}

func NewInitializeCommand() Command {
	return Command{
		commandType: INITIALIZE,
		data:        nil,
	}
}

type Move struct {
	Side       Player
	Coordinate Coordinate
}

func NewMoveCommand(side Player, coordinate Coordinate) Command {
	return Command{
		commandType: MOVE,
		data: Move{
			Side:       side,
			Coordinate: coordinate,
		},
	}
}

type CommandRejectHandler interface {
	InvalidCommand(command Command)
}

type CommandPolicy interface {
	processCommand(command Command, rejectHandler CommandRejectHandler)
	ofType() string
}
type CommandHandler struct {
	eventConsumer EventConsumer
	commandPolicy CommandPolicy
}

type UninitializedGameCommandPolicy struct {
	eventConsumer EventConsumer
}

func (policy UninitializedGameCommandPolicy) processCommand(command Command, rejectHandler CommandRejectHandler) {
	if command.commandType != INITIALIZE {
		rejectHandler.InvalidCommand(command)
	}

	policy.eventConsumer.SendEvent(NewInitializedEvent())
}
func (policy UninitializedGameCommandPolicy) ofType() string {
	return "UninitializedGameCommandPolicy"
}

type InProgressCommandPolicy struct {
	eventConsumer EventConsumer
	gameState     GameState
}

func (policy InProgressCommandPolicy) processCommand(command Command, rejectHandler CommandRejectHandler) {
	if command.commandType == MOVE {
		move := command.data.(Move)
		side := move.Side
		coordinate := move.Coordinate

		gameState := policy.gameState
		if side == gameState.PlayerTurn {
			if gameState.PossibleMoves.moves[coordinate] {
				policy.eventConsumer.SendEvent(NewMoveEvent(coordinate))
				return
			}
		}
	}
	rejectHandler.InvalidCommand(command)
}
func (policy InProgressCommandPolicy) ofType() string {
	return "InProgressCommandPolicyy"
}

func (commandHandler *CommandHandler) AttemptCommand(command Command, rejectHandler CommandRejectHandler) {
	commandHandler.commandPolicy.processCommand(command, rejectHandler)
}

func (commandHandler *CommandHandler) StateUpdated(gameState GameState) {
	commandHandler.commandPolicy = InProgressCommandPolicy{
		eventConsumer: commandHandler.eventConsumer,
		gameState:     gameState,
	}
}

func NewCommandHandler(eventConsumer EventConsumer, notifier StateUpdateSource) *CommandHandler {
	commandHandler := CommandHandler{
		eventConsumer: eventConsumer,
		commandPolicy: UninitializedGameCommandPolicy{eventConsumer: eventConsumer},
	}

	notifier.Register(&commandHandler)

	return &commandHandler
}
