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
			fmt.Printf("x -> %d, y -> %d\n", dx, dy)
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

	fmt.Printf("(%d, %d) is in bounds\n", coordinate.X, coordinate.Y)
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

func IsPossibleMove(side Player, possibleMove Coordinate, board map[Coordinate]CellClaim) bool {
	if !inBounds(possibleMove) {
		return false
	}
	if isOccupied(possibleMove, board) {
		return false
	}

	directions := make([]Direction, 8)
	directionCount := 0
	neighborhood := getNeighborhood(possibleMove)
	for _, cell := range neighborhood {
		x := cell.X
		y := cell.Y

		if inBounds(cell) && board[cell] != nil && board[cell].OwnedBy(side.opposite()) {
			directions[directionCount] = Direction{
				X: x - possibleMove.X,
				Y: y - possibleMove.Y,
			}
			directionCount = directionCount + 1
		}
	}

	directions = directions[0:directionCount]

	atLeastOneDirectionResolves := false
	for _, direction := range directions {
		atLeastOneDirectionResolves = DoesDirectionResolve(side, possibleMove, direction, board)
		fmt.Printf("Checked direction (%d, %d) and got resolves = %t\n", direction.X, direction.Y, atLeastOneDirectionResolves)
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
				fmt.Printf("Part of the edge -> (%d, %d)\n", v.X, v.Y)
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
		fmt.Printf("%d", count)
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

func (brain *GameBrain) printGameState() {
	gameState := brain.GameState
	board := gameState.board
	edge := gameState.edge
	possibleMoves := gameState.possibleMoves

	fmt.Println()
	bounds := sequence(0, 8)
	for x := range bounds {
		rowString := ""
		for y := range bounds {
			next := "?"
			coordinate := Coordinate{X: x, Y: y}
			owner := board[coordinate]

			if possibleMoves.moves[coordinate] {
				next = "_"
			} else if edge[coordinate] {
				next = "e"
			} else if owner == nil {
				next = "-"
			} else if owner.OwnedBy(BLACK) {
				next = "B"
			} else if owner.OwnedBy(WHITE) {
				next = "W"
			}

			rowString = rowString + next
		}
		fmt.Println(rowString)
	}
}

func NewGameBrain() GameBrain {
	gameBrain := GameBrain{
		GameState: getInitialGameState(),
	}

	gameBrain.printGameState()

	return gameBrain
}
