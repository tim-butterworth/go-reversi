import * as r from "rambda"

import {
    Maybe,
    some,
    none
} from "./maybe"

type AccessKey = string | number;
type Result<T> = object | undefined | T;
const safelyGet = <T>(obj: object, accessors: AccessKey[]): Result<T> => {
    let value = obj
    const max = accessors.length

    let index = 0
    let keepGoing = true
    while (keepGoing) {
        if (!value) {
            keepGoing = false
        } else {
            value = value[accessors[index]]

            index++
            keepGoing = index < max
        }
    }

    return value
}

enum Sides {
    WHITE = "WHITE",
    BLACK = "BLACK",
    NOT_ASSIGNED = "NOT_ASSIGNED",
}
enum GameStates {
    NOT_STARTED = "NOT_STARTED",
    STARTED = "STARTED",
    FINISHED = "FINISHED"
}

interface StateOfBoard<T extends GameStates> {
    gameState: T
}
interface PendingGame extends StateOfBoard<GameStates.NOT_STARTED> { }
interface InProgressGame extends StateOfBoard<GameStates.STARTED> {
    edge: Coordinate[];
    used: { [key: number]: { [key: number]: Sides } };
    availableMoves: Coordinate[];
    playerTurn: Sides;
    showMoves: boolean;
}
interface FinishedGame extends StateOfBoard<GameStates.FINISHED> {
    used: { [key: number]: { [key: number]: Sides } };
}
type BoardState = PendingGame | InProgressGame | FinishedGame;
type AppState = {
    side: Sides,
    isComitting: boolean,
    boardState: BoardState,
    events: actions[],
    restoreState: BoardState[],
    undoCount: number,
    pendingMove: Maybe<Coordinate>,
}

const initialState: AppState = {
    side: Sides.NOT_ASSIGNED,
    isComitting: false,
    boardState: {
        gameState: GameStates.NOT_STARTED,
    },
    restoreState: [],
    undoCount: 0,
    events: [],
    pendingMove: none()
}

type Coordinate = {
    X: number,
    Y: number,
}
const opposite = (side: Sides.WHITE | Sides.BLACK): Sides.WHITE | Sides.BLACK => {
    if (side === Sides.WHITE) {
        return Sides.BLACK
    } else {
        return Sides.WHITE
    }
}

const inBounds = ({ X, Y }, { dx, dy }) => {
    const nextX = X + dx;
    const nextY = Y + dy;

    return (0 <= nextX && nextX < 8) && (0 <= nextY && nextY < 8)
}

const calculateFlips = (toFlip: Sides.WHITE | Sides.BLACK, used: { [key: number]: { [key: number]: Sides } }, move: Coordinate) => {
    const directions: Direction[] = [
        { dx: 1, dy: 0 },
        { dx: 0, dy: 1 },
        { dx: -1, dy: 0 },
        { dx: 0, dy: -1 },
        { dx: 1, dy: 1 },
        { dx: -1, dy: 1 },
        { dx: 1, dy: -1 },
        { dx: -1, dy: -1 },
    ].filter((direction) => inBounds(move, direction))

    const edgeSide = opposite(toFlip)
    return directions
        .map((direction) => collectFlips({ move, direction, flipSide: toFlip, edgeSide, used }))
        .reduce((accume, v) => [...accume, ...v], [])
}

type Direction = {
    dx: number;
    dy: number;
}
type CollectFlipData = {
    move: Coordinate;
    direction: Direction;
    flipSide: Sides.BLACK | Sides.WHITE;
    edgeSide: Sides.BLACK | Sides.WHITE;
    used: { [key: number]: { [key: number]: Sides } }
}
const collectFlips = ({ move, direction, flipSide, edgeSide, used }: CollectFlipData): Coordinate[] => {
    const step = ({ X, Y }, { dx, dy }) => ({ X: X + dx, Y: Y + dy })
    let current = move

    const bucket = [];
    let found = false;
    let keepGoing = true;
    while (keepGoing) {
        current = step(current, direction)
        const valueAtCurrent = safelyGet<Sides.BLACK | Sides.WHITE>(used, [current.X, current.Y])
        if (valueAtCurrent === edgeSide) {
            keepGoing = false;
            found = true;
        } else if (valueAtCurrent === flipSide) {
            bucket.push(current);
        } else {
            keepGoing = false;
        }
    }

    if (found) {
        return bucket;
    } else {
        return [];
    }
}

const flip = (
    toFlip: Coordinate[],
    current: { [key: number]: { [key: number]: Sides } }
): { [key: number]: { [key: number]: Sides } } => {
    const toFlipLookup: { [key: number]: { [key: number]: string } } = toFlip.reduce(
        (accume, { X, Y }) => {
            const column = accume[X] || {};
            column[Y] = "flip"
            accume[X] = column

            return accume;
        },
        {}
    )

    const result = {};
    const xs = Object.keys(current);
    xs.forEach((x) => {
        const ys = Object.keys(current[x]);
        ys.forEach((y) => {
            const column = result[x] || {};
            if (safelyGet<string>(toFlipLookup, [x, y]) === "flip") {
                column[y] = opposite(current[x][y])
            } else {
                column[y] = current[x][y];
            }

            result[x] = column;
        })
    })

    return result;
}

const getNeighborhood = (coordinate: Coordinate): Coordinate[] => {
    const neighborhood: Coordinate[] = [];
    const modifiers = [-1, 0, 1];
    const { X, Y } = coordinate;
    modifiers.forEach((dx) => {
        modifiers.forEach((dy) => {
            if (inBounds(coordinate, { dx, dy })) {
                neighborhood.push({ X: (X + dx), Y: (Y + dy) })
            }
        })
    })

    return neighborhood;
}

const updateEdge = (edge: Coordinate[], used: { [key: number]: { [key: number]: Sides } }, move: Coordinate): Coordinate[] => {
    const neighbors: Coordinate[] = getNeighborhood(move);
    const updatedEdge: Coordinate[] = edge.filter(({ X, Y }) => !(X === move.X && Y === move.Y));
    const edgeLookup: { [key: number]: { [key: number]: string } } = updatedEdge.reduce(
        (accume, { X, Y }) => {
            const existing = (accume[X] || {});
            accume[X] = existing;
            existing[Y] = "used"

            return accume;
        },
        {}
    )
    const unusedNeighbors = neighbors
        .filter(({ X, Y }) => !(r.path([X, Y], used)))
        .filter(({ X, Y }) => !(r.path([X, Y], edgeLookup)))

    unusedNeighbors.forEach((coordinate) => {
        updatedEdge.push(coordinate)
    })

    return updatedEdge;
}

const getMoves = (
    playerTurn: Sides.BLACK | Sides.WHITE,
    edge: Coordinate[],
    used: { [key: number]: { [key: number]: Sides } }
): Coordinate[] => {
    const availableMoves: Coordinate[] = [];
    for (const move of edge) {
        const flips = calculateFlips(opposite(playerTurn), used, move)
        if (flips.length > 0) {
            availableMoves.push(move);
        }
    }

    return availableMoves;
}

type UpdatedMovesRequest = {
    playerTurn: Sides.BLACK | Sides.WHITE;
    edge: Coordinate[];
    used: { [key: number]: { [key: number]: Sides } };
}
type UpdatedMovesResponse = {
    updatedPlayerTurn: Sides.BLACK | Sides.WHITE;
    availableMoves: Coordinate[];
}
const getUpdatedMovesAndPlayer = ({
    playerTurn,
    edge,
    used
}: UpdatedMovesRequest): UpdatedMovesResponse => {
    const otherPlayer = opposite(playerTurn);
    const samePlayer = playerTurn;

    let updatedPlayerTurn: Sides.BLACK | Sides.WHITE
    let availableMoves = getMoves(otherPlayer, edge, used);
    if (availableMoves.length > 0) {
        updatedPlayerTurn = otherPlayer
    } else {
        availableMoves = getMoves(samePlayer, edge, used)
        updatedPlayerTurn = samePlayer;
    }

    return {
        updatedPlayerTurn,
        availableMoves
    }
}

const updateBoardState = ({ edge, used, playerTurn, showMoves }, move: Coordinate): InProgressGame => {
    let updatedUsed: { [key: number]: { [key: number]: Sides } }
    if (playerTurn === Sides.BLACK) {
        const toFlip = calculateFlips(Sides.WHITE, used, move)
        updatedUsed = flip(toFlip, used)

        const existing = updatedUsed[move.X] || {};
        updatedUsed[move.X] = existing;
        existing[move.Y] = playerTurn;
    } else {
        const toFlip = calculateFlips(Sides.BLACK, used, move)
        updatedUsed = flip(toFlip, used)

        const existing = updatedUsed[move.X] || {};
        updatedUsed[move.X] = existing;
        existing[move.Y] = playerTurn;
    }

    const updatedEdge = updateEdge(edge, updatedUsed, move)

    const { updatedPlayerTurn, availableMoves } = getUpdatedMovesAndPlayer({
        playerTurn,
        edge: updatedEdge,
        used: updatedUsed
    })

    return {
        edge: updatedEdge,
        used: updatedUsed,
        playerTurn: updatedPlayerTurn,
        availableMoves,
        gameState: GameStates.STARTED,
        showMoves
    }
}

enum ActionTypes {
    SIDE_ASSIGNED = "SIDE_ASSIGNED",
    INITIALIZED = "INITIALIZED",
    MOVED = "MOVED",
    SHOW_MOVES = "SHOW_MOVES",
    HIDE_MOVES = "HIDE_MOVES",
    PREVIEW_MOVE = "PREVIEW_MOVE",
    UNDO = "UNDO",
    COMMIT_MOVE = "COMMIT_MOVE",
    COMMIT = "COMMIT",
    MOVE_ACCEPTED = "MOVE_ACCEPTED",
    MOVE_REJECTED = "MOVE_REJECTED"
}

interface Action<T extends ActionTypes> {
    actionType: T
}
interface SideAssigned extends Action<ActionTypes.SIDE_ASSIGNED> {
    data: Sides.BLACK | Sides.WHITE;
}
const getSideAssigned = (side: Sides.BLACK | Sides.WHITE): SideAssigned => ({
    actionType: ActionTypes.SIDE_ASSIGNED,
    data: side
})

interface Initialized extends Action<ActionTypes.INITIALIZED> { }
const getInitialized = (): Initialized => ({
    actionType: ActionTypes.INITIALIZED
})

interface Moved extends Action<ActionTypes.MOVED> {
    data: Coordinate
}
const getMoved = (data: Coordinate): Moved => ({
    actionType: ActionTypes.MOVED,
    data
})

interface ShowMoves extends Action<ActionTypes.SHOW_MOVES> { }
const getShow = (): ShowMoves => ({
    actionType: ActionTypes.SHOW_MOVES
})

interface HideMoves extends Action<ActionTypes.HIDE_MOVES> { }
const getHide = (): HideMoves => ({
    actionType: ActionTypes.HIDE_MOVES
})

interface PreviewMove extends Action<ActionTypes.PREVIEW_MOVE> {
    data: Coordinate
}
const getPreviewMove = (data: Coordinate): PreviewMove => ({
    actionType: ActionTypes.PREVIEW_MOVE,
    data
})
interface UndoMove extends Action<ActionTypes.UNDO> { }
const getUndo = (): UndoMove => ({
    actionType: ActionTypes.UNDO
})
interface CommitMove extends Action<ActionTypes.COMMIT> {
    data: Coordinate;
}
const getCommit = (data: Coordinate): CommitMove => ({
    actionType: ActionTypes.COMMIT,
    data
})
interface MoveAccepted extends Action<ActionTypes.MOVE_ACCEPTED> { }
const getMoveAccepted = (): MoveAccepted => ({
    actionType: ActionTypes.MOVE_ACCEPTED
})

interface MoveRejected extends Action<ActionTypes.MOVE_REJECTED> { }
const getMoveRejected = (): MoveRejected => ({
    actionType: ActionTypes.MOVE_REJECTED
})

type actions = SideAssigned
    | Initialized
    | Moved
    | ShowMoves
    | HideMoves
    | PreviewMove
    | UndoMove
    | CommitMove
    | MoveRejected
    | MoveAccepted;

const handleAction = ({ action, state }: { action: actions, state: AppState }): AppState => {
    if (action.actionType === ActionTypes.SIDE_ASSIGNED) {
        return {
            ...state,
            side: action.data
        }
    }

    if (action.actionType === ActionTypes.INITIALIZED) {
        const edge = [
            { X: 2, Y: 2 },
            { X: 2, Y: 3 },
            { X: 2, Y: 4 },
            { X: 2, Y: 5 },
            { X: 5, Y: 2 },
            { X: 5, Y: 3 },
            { X: 5, Y: 4 },
            { X: 5, Y: 5 },
            { X: 3, Y: 2 },
            { X: 4, Y: 2 },
            { X: 3, Y: 5 },
            { X: 4, Y: 5 },
        ];
        const used = {
            3: {
                3: Sides.BLACK,
                4: Sides.WHITE
            },
            4: {
                4: Sides.BLACK,
                3: Sides.WHITE
            }
        };
        const playerTurn = Sides.BLACK;
        const gameState = GameStates.STARTED;
        const availableMoves = getMoves(playerTurn, edge, used);

        return {
            ...state,
            boardState: {
                gameState,
                edge,
                used,
                availableMoves,
                playerTurn,
                showMoves: true
            }
        }
    }

    if (state.boardState.gameState === GameStates.STARTED) {
        if (action.actionType === ActionTypes.MOVED) {
            const updatedBoardState = updateBoardState(state.boardState, action.data)
            return {
                ...state,
                boardState: updatedBoardState
            }
        }

        if (action.actionType === ActionTypes.SHOW_MOVES || action.actionType === ActionTypes.HIDE_MOVES) {
            return {
                ...state,
                boardState: {
                    ...(state.boardState),
                    showMoves: (action.actionType === ActionTypes.SHOW_MOVES)
                }
            }
        }

        if (action.actionType === ActionTypes.PREVIEW_MOVE && state.undoCount <= 0) {
            const coordinate: Coordinate = action.data;
            if (r.includes(coordinate, state.boardState.availableMoves)) {
                const updatedBoardState = updateBoardState(state.boardState, coordinate)
                return {
                    ...state,
                    boardState: updatedBoardState,
                    restoreState: [state.boardState, ...state.restoreState],
                    undoCount: (state.undoCount + 1),
                    pendingMove: some<Coordinate>(coordinate)
                }
            }
        }

        if (action.actionType === ActionTypes.COMMIT) {
            return {
                ...state,
                isComitting: true
            }
        }

        if (action.actionType === ActionTypes.MOVE_REJECTED && state.isComitting) {
            const [restore, ...restoreState] = state.restoreState;
            return {
                ...state,
                isComitting: false,
                boardState: restore,
                restoreState,
                undoCount: (state.undoCount - 1),
            }
        }

        if (action.actionType === ActionTypes.MOVE_ACCEPTED) {
            return {
                ...state,
                isComitting: false,
                restoreState: [],
                undoCount: 0,
                pendingMove: none()
            }
        }

        if (action.actionType === ActionTypes.UNDO && state.undoCount > 0) {
            const [restore, ...restoreState] = state.restoreState;
            return {
                ...state,
                boardState: restore,
                restoreState,
                undoCount: (state.undoCount - 1),
                pendingMove: none()
            }
        }
    }
    return state;
}

const reducer = ({ action, state = initialState }: { action: actions, state: AppState }): AppState => {
    let updatedState: AppState = {
        ...state,
        events: [...(state.events), action]
    }

    return handleAction({ state: updatedState, action })
}

export {
    reducer,
    getSideAssigned,
    getInitialized,
    getMoved,
    getShow,
    getHide,
    getPreviewMove,
    getUndo,
    getCommit,
    getMoveAccepted,
    getMoveRejected,
    Sides,
    GameStates,
    AppState,
    Coordinate
}
