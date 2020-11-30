package core

type Player string

const (
	BLACK Player = "BLACK"
	WHITE Player = "WHITE"
)

func (player Player) String() string {
	switch player {
	case BLACK:
		return "BLACK"
	case WHITE:
		return "WHITE"
	default:
		return "UNKNOWN"
	}
}

type EventType string

const (
	INITILIZED EventType = "INITIALIZED"
	MOVED      EventType = "MOVED"
)

type Event struct {
	EventType EventType
	Data      interface{}
}

func NewInitializedEvent() Event {
	return Event{
		EventType: INITILIZED,
		Data:      nil,
	}
}

func NewMoveEvent(coordinate Coordinate) Event {
	return Event{
		EventType: MOVED,
		Data:      coordinate,
	}
}

type CellClaim interface {
	OwnedBy(player Player) bool
}
type Coordinate struct {
	X int
	Y int
}
type Direction = Coordinate
type ownedByBlack struct{}

func (owner ownedByBlack) OwnedBy(player Player) bool {
	return player == BLACK
}

type ownedByWhite struct{}

func (owner ownedByWhite) OwnedBy(player Player) bool {
	return player == WHITE
}

type notOwned struct{}

func (owner notOwned) OwnedBy(player Player) bool {
	return false
}

type EventConsumer interface {
	SendEvent(event Event)
}

type StateUpdateConsumer interface {
	StateUpdated(gameState GameState)
}
type StateUpdateSource interface {
	Register(consumer StateUpdateConsumer)
}

func GetInitialBoard() map[Coordinate]CellClaim {
	result := make(map[Coordinate]CellClaim)

	ownedByBlack := ownedByBlack{}
	ownedByWhite := ownedByWhite{}

	result[Coordinate{X: 3, Y: 3}] = ownedByBlack
	result[Coordinate{X: 4, Y: 4}] = ownedByBlack

	result[Coordinate{X: 4, Y: 3}] = ownedByWhite
	result[Coordinate{X: 3, Y: 4}] = ownedByWhite

	return result
}

func collectUsed(board map[Coordinate]CellClaim) map[Coordinate]bool {
	used := make(map[Coordinate]bool)

	for coordinate := range board {
		used[coordinate] = true
	}

	return used
}

func getNeighborhood(c Coordinate) []Coordinate {
	result := make([]Coordinate, 8)

	x := c.X
	y := c.Y

	values := []int{-1, 0, 1}

	itterator := NewCounter([][]int{values, values})
	index := 0
	for {
		if !itterator.HasNext() {
			break
		}

		v := itterator.Next()
		dx := v[0]
		dy := v[1]

		if !(dx == 0 && dy == 0) {
			result[index] = Coordinate{X: x + dx, Y: y + dy}
			index = index + 1
		}
	}

	return result
}

func collectEdge(board map[Coordinate]CellClaim, used map[Coordinate]bool) map[Coordinate]bool {
	edge := make(map[Coordinate]bool)

	for k := range used {
		neighborhood := getNeighborhood(k)
		for _, v := range neighborhood {
			if !used[v] {
				edge[v] = true
			}
		}
	}

	return edge
}

func getInitialGameState() GameState {
	initialSide := BLACK

	board := GetInitialBoard()
	used := collectUsed(board)
	edge := collectEdge(board, used)
	possibleMoves := possibleMovesFor(edge, board, initialSide)

	gameState := GameState{
		Board:         board,
		PlayerTurn:    initialSide,
		Used:          used,
		Edge:          edge,
		PossibleMoves: possibleMoves,
	}

	return gameState
}

func DoesDirectionResolve(side Player, move Coordinate, direction Direction, board map[Coordinate]CellClaim) bool {
	location := step(move, direction)

	resolves := false
	crossedOppositeOwned := false
	for {
		if !inBounds(location) {
			break
		}
		if !isOccupied(location, board) {
			break
		}
		if ownClaim(side, location, board) {
			resolves = crossedOppositeOwned
			break
		}
		if oppositeClaim(side, location, board) {
			crossedOppositeOwned = true
		}

		location = step(location, direction)
	}
	return resolves
}

func possibleDirections(side Player, move Coordinate, board map[Coordinate]CellClaim) []Direction {
	directions := make([]Direction, 8)
	directionCount := 0
	neighborhood := getNeighborhood(move)
	for _, cell := range neighborhood {
		x := cell.X
		y := cell.Y

		if inBounds(cell) && board[cell] != nil && board[cell].OwnedBy(side.opposite()) {
			directions[directionCount] = Direction{
				X: x - move.X,
				Y: y - move.Y,
			}
			directionCount = directionCount + 1
		}
	}

	return directions[0:directionCount]
}

func IsPossibleMove(side Player, possibleMove Coordinate, board map[Coordinate]CellClaim) bool {
	if !inBounds(possibleMove) {
		return false
	}
	if isOccupied(possibleMove, board) {
		return false
	}

	directions := possibleDirections(side, possibleMove, board)

	atLeastOneDirectionResolves := false
	for _, direction := range directions {
		atLeastOneDirectionResolves = DoesDirectionResolve(side, possibleMove, direction, board)
		if atLeastOneDirectionResolves {
			break
		}
	}

	return atLeastOneDirectionResolves
}

func possibleMovesFor(edge map[Coordinate]bool, board map[Coordinate]CellClaim, side Player) possibleMoves {
	moves := make(map[Coordinate]bool)
	count := 0
	for e := range edge {
		if IsPossibleMove(side, e, board) {
			moves[e] = true
		}
		count = count + 1
	}

	return possibleMoves{
		side:  side,
		moves: moves,
	}
}

func (player Player) opposite() Player {
	if player == WHITE {
		return BLACK
	}

	return WHITE
}

func matchingSide(side Player, coordinate Coordinate, board map[Coordinate]CellClaim) bool {
	claim := board[coordinate]

	if claim == nil {
		return false
	}

	return claim.OwnedBy(side)
}

func oppositeClaim(side Player, coordinate Coordinate, board map[Coordinate]CellClaim) bool {
	return matchingSide(side.opposite(), coordinate, board)
}

func ownClaim(side Player, coordinate Coordinate, board map[Coordinate]CellClaim) bool {
	return matchingSide(side, coordinate, board)
}

func inBounds(coordinate Coordinate) bool {
	x := coordinate.X
	if x < 0 || x >= 8 {
		return false
	}

	y := coordinate.Y
	if y < 0 || y >= 8 {
		return false
	}

	return true
}

func isOccupied(coordinate Coordinate, board map[Coordinate]CellClaim) bool {
	claim := board[coordinate]

	if claim == nil {
		return false
	}

	return claim.OwnedBy(WHITE) || claim.OwnedBy(BLACK)
}

func step(move Coordinate, direction Direction) Coordinate {
	return Coordinate{
		X: move.X + direction.X,
		Y: move.Y + direction.Y,
	}
}

func cellsToFlipInThisDirection(direction Direction, start Coordinate, board map[Coordinate]CellClaim, side Player) []Coordinate {
	result := make([]Coordinate, 8)
	toFlipCount := 0

	location := step(start, direction)

	pendingToFlipCount := 0
	foundSomeToFlip := false
	for {
		if !inBounds(location) {
			break
		}
		if !isOccupied(location, board) {
			break
		}
		if ownClaim(side, location, board) {
			if foundSomeToFlip {
				toFlipCount = pendingToFlipCount
			}
			break
		}
		if oppositeClaim(side, location, board) {
			foundSomeToFlip = true
		}

		result[pendingToFlipCount] = Coordinate{
			X: location.X,
			Y: location.Y,
		}

		pendingToFlipCount++
		location = step(location, direction)
	}

	return result[0:toFlipCount]
}

func updateEdge(edge map[Coordinate]bool, used map[Coordinate]bool, coordinate Coordinate) map[Coordinate]bool {
	delete(edge, coordinate)
	neighborhood := getNeighborhood(coordinate)

	for _, neighbor := range neighborhood {
		if !used[neighbor] {
			edge[neighbor] = true
		}
	}

	return edge
}

func addAll(m map[Coordinate]bool, flipList []Coordinate) {
	for _, coordinate := range flipList {
		m[coordinate] = true
	}
}

func getCellsToFlip(board map[Coordinate]CellClaim, coordinate Coordinate, side Player) map[Coordinate]bool {
	result := make(map[Coordinate]bool)

	directions := possibleDirections(side, coordinate, board)

	for _, direction := range directions {
		addAll(result, cellsToFlipInThisDirection(direction, coordinate, board, side))
	}

	return result
}

func applyMove(gameState GameState, side Player, coordinate Coordinate) GameState {
	var owner CellClaim
	if side == BLACK {
		owner = ownedByBlack{}
	} else {
		owner = ownedByWhite{}
	}

	cellsToFlip := getCellsToFlip(gameState.Board, coordinate, side)
	for cell := range cellsToFlip {
		gameState.Board[cell] = owner
	}
	gameState.Board[coordinate] = owner

	gameState.Used[coordinate] = true
	gameState.Edge = updateEdge(gameState.Edge, gameState.Used, coordinate)

	// Try to get moves for the opposite side
	possibleMoves := possibleMovesFor(gameState.Edge, gameState.Board, side.opposite())
	if len(possibleMoves.moves) == 0 {
		// If there are no moves for the opposite side, get moves for the same side
		possibleMoves = possibleMovesFor(gameState.Edge, gameState.Board, side)
	}

	gameState.PossibleMoves = possibleMoves
	gameState.PlayerTurn = possibleMoves.side

	return gameState
}

type gameEventAggregator struct {
	state                GameState
	stateUpdateConsumers []StateUpdateConsumer
}

func (aggregator *gameEventAggregator) SendEvent(event Event) {
	if event.EventType == INITILIZED {
		aggregator.state = getInitialGameState()
	}

	if event.EventType == MOVED {
		side := aggregator.state.PlayerTurn
		coordinate := event.Data.(Coordinate)

		aggregator.state = applyMove(aggregator.state, side, coordinate)
	}

	for _, consumer := range aggregator.stateUpdateConsumers {
		consumer.StateUpdated(aggregator.state)
	}
}

func (aggregator *gameEventAggregator) Register(consumer StateUpdateConsumer) {
	consumers := aggregator.stateUpdateConsumers
	updatedSize := len(consumers) + 1
	updatedConsumers := make([]StateUpdateConsumer, updatedSize)
	for i, consumer := range consumers {
		updatedConsumers[i] = consumer
	}

	updatedConsumers[updatedSize-1] = consumer
	aggregator.stateUpdateConsumers = updatedConsumers
}

type possibleMoves struct {
	side  Player
	moves map[Coordinate]bool
}
type GameState struct {
	Board         map[Coordinate]CellClaim
	PlayerTurn    Player
	Used          map[Coordinate]bool
	Edge          map[Coordinate]bool
	PossibleMoves possibleMoves
}

func (gameState GameState) MoveOptions() map[Coordinate]bool {
	return gameState.PossibleMoves.moves
}

type StateUpdaterAndEventConsumer interface {
	StateUpdateSource
	EventConsumer
}

func NewGameEventAggregator() StateUpdaterAndEventConsumer {
	return &gameEventAggregator{stateUpdateConsumers: []StateUpdateConsumer{}}
}
