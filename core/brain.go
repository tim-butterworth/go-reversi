package core

import (
	"fmt"
)

type Responder interface {
	Respond(message string)
	RespondAll(message string)
	NotifyActivePlayer(message string, activePlayerSide Player)
	NotifyInactivePlayer(message string, activePlayerSide Player)
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

type GameBrain struct {
	GameState gameState
}

func (brain *GameBrain) ExecuteCommand(command GameCommand, responder Responder) {
	responder.RespondAll("Hi from the brain of the game!")
	responder.Respond("Special secret message")

	responder.NotifyActivePlayer("Your turn!", brain.GameState.playerTurn)
	responder.NotifyInactivePlayer("Not your turn just yet!", brain.GameState.playerTurn)
}

func stringValue(b bool) string {
	if b {
		return "t"
	}
	return "f"
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

func (player Player) opposite() Player {
	if player == WHITE {
		return BLACK
	}

	return WHITE
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

type possibleMoves struct {
	side  Player
	moves map[Coordinate]bool
}
type gameState struct {
	board         map[Coordinate]CellClaim
	playerTurn    Player
	used          map[Coordinate]bool
	edge          map[Coordinate]bool
	possibleMoves possibleMoves
}

func sequence(min int, max int) []int {
	if max <= min {
		return make([]int, 0)
	}

	result := make([]int, max-min)
	current := min
	index := 0
	for {
		if !(current < max) {
			break
		}

		result[index] = current

		index++
		current = min + index
	}

	return result
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

func getPossibleMoves(edge map[Coordinate]bool, board map[Coordinate]CellClaim, side Player) possibleMoves {
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

func getInitialGameState() gameState {
	initialSide := BLACK

	board := GetInitialBoard()
	used := collectUsed(board)
	edge := collectEdge(board, used)
	possibleMoves := getPossibleMoves(edge, board, initialSide)

	gameState := gameState{
		board:         board,
		playerTurn:    initialSide,
		used:          used,
		edge:          edge,
		possibleMoves: possibleMoves,
	}

	return gameState
}

func (brain *GameBrain) PrintGameState() {
	gameState := brain.GameState
	board := gameState.board
	edge := gameState.edge
	possibleMoves := gameState.possibleMoves

	fmt.Println()
	bounds := sequence(0, 8)
	for y := range bounds {
		rowString := ""
		for x := range bounds {
			next := "[?]"
			coordinate := Coordinate{X: x, Y: y}
			owner := board[coordinate]

			if possibleMoves.moves[coordinate] {
				next = "[ ]"
			} else if edge[coordinate] {
				next = " e "
			} else if owner == nil {
				next = "[-]"
			} else if owner.OwnedBy(BLACK) {
				next = "[B]"
			} else if owner.OwnedBy(WHITE) {
				next = "[W]"
			}

			rowString = rowString + next
		}
		fmt.Println(rowString)
	}
}

type NextPlayInfo struct {
	player         Player
	availableMoves []Coordinate
}
type MoveSuccessResult struct {
	nextPlay    NextPlayInfo
	appliedMove Move
}

func (nextPlayerInfo NextPlayInfo) NextPlayer() Player {
	return nextPlayerInfo.player
}
func (nextPlayerInfo NextPlayInfo) Moves() []Coordinate {
	return nextPlayerInfo.availableMoves
}
func (result MoveSuccessResult) NextPlayerInfo() NextPlayInfo {
	return result.nextPlay
}

type ResultHandler interface {
	MoveSuccess(result MoveSuccessResult)
	GameInitialized(nextPlay NextPlayInfo)
	MoveFailure()
}

type Move struct {
	Side       Player
	Coordinate Coordinate
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

func nextPlayInfoFromGameState(gameState gameState) NextPlayInfo {
	movesThatArePossible := gameState.possibleMoves.moves
	availableMoves := make([]Coordinate, len(movesThatArePossible))
	index := 0
	for move := range movesThatArePossible {
		availableMoves[index] = move
		index++
	}

	return NextPlayInfo{
		player:         gameState.playerTurn,
		availableMoves: availableMoves,
	}
}

func getMoveResult(appliedMove Move, gameState gameState) MoveSuccessResult {
	return MoveSuccessResult{
		nextPlay:    nextPlayInfoFromGameState(gameState),
		appliedMove: appliedMove,
	}
}

func (brain *GameBrain) AttemptMove(move Move, resultHandler ResultHandler) {
	side := move.Side
	if brain.GameState.playerTurn != side {
		resultHandler.MoveFailure()
		return
	}

	coordinate := move.Coordinate
	if brain.GameState.used[coordinate] {
		resultHandler.MoveFailure()
		return
	}

	if !brain.GameState.possibleMoves.moves[coordinate] {
		resultHandler.MoveFailure()
		return
	}

	var owner CellClaim
	if side == BLACK {
		owner = ownedByBlack{}
	} else {
		owner = ownedByWhite{}
	}

	gameState := brain.GameState

	cellsToFlip := getCellsToFlip(gameState.board, coordinate, side)
	for cell := range cellsToFlip {
		gameState.board[cell] = owner
	}
	gameState.board[coordinate] = owner

	gameState.used[coordinate] = true
	gameState.edge = updateEdge(gameState.edge, gameState.used, coordinate)

	// Try to get moves for the opposite side
	possibleMoves := getPossibleMoves(gameState.edge, gameState.board, side.opposite())
	if len(possibleMoves.moves) == 0 {
		// If there are no moves for the opposite side, get moves for the same side
		possibleMoves = getPossibleMoves(gameState.edge, gameState.board, side)
	}

	gameState.possibleMoves = possibleMoves
	gameState.playerTurn = possibleMoves.side
	brain.GameState = gameState

	moveResult := getMoveResult(move, brain.GameState)
	resultHandler.MoveSuccess(moveResult)
}

func (brain *GameBrain) Initialize(resultHandler ResultHandler) {
	resultHandler.GameInitialized(nextPlayInfoFromGameState(brain.GameState))
}

func NewGameBrain(resultHandler ResultHandler) GameBrain {
	return GameBrain{GameState: getInitialGameState()}
}
