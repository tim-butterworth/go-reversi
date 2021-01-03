import { getStore } from "./store"
import {
    reducer,
    getSideAssigned,
    getInitialized,
    getMoved,
    getShow,
    getHide,
    Sides,
    GameStates,
    AppState
} from "./reducer"

import * as r from "rambda"

const socket = new WebSocket(`ws://${location.host}/ws`)

const withAction = (rawEvent, actionConsumer) => {
    try {
        const unmappedAction = JSON.parse(JSON.parse(rawEvent).Text)
        if (unmappedAction.EventType === "SIDE_ASSIGNED") {
            actionConsumer(getSideAssigned(unmappedAction.Data.Side))
        }

        if (unmappedAction.EventType === "INITIALIZED") {
            actionConsumer(getInitialized())
        }

        if (unmappedAction.EventType === "MOVED") {
            actionConsumer(getMoved(unmappedAction.Data))
        }
    } catch (error) {
        console.log(`got a json parsing error attempting to parse ${rawEvent} -> ${error}`)
    }
}

const ifClaimedOrElse = ({ data, ifClaimed, orElse }) => {
    const { claimed, coordinate: { x, y } } = data;

    const claimer = r.path([x, y], claimed)

    if (claimer === Sides.WHITE) {
        ifClaimed("white");
    } else if (claimer === Sides.BLACK) {
        ifClaimed("black");
    } else {
        orElse();
    }
}

const range = (min: number, max: number) => {
    const result = []
    let current = min
    while (current < max) {
        result.push(current)
        current++
    }

    return result
}

const getShowHideColors = ({ showMoves }): { showColor: string; hideColor: string; } => {
    const activeColor = "red";
    const inactiveColor = "blue";

    if (showMoves) {
        return {
            showColor: activeColor,
            hideColor: inactiveColor
        }
    } else {
        return {
            showColor: inactiveColor,
            hideColor: activeColor
        }
    }
}

type CountKeys = keyof typeof Sides;

const getCounts = (used: { [key: number]: { [key: number]: Sides } }): { [key in CountKeys]: number } => {
    const result: { [key in CountKeys]: number } = {
        "WHITE": 0,
        "BLACK": 0,
        "NOT_ASSIGNED": 0
    }

    const xs = Object.keys(used);
    for (const x of xs) {
        const ys = Object.keys(used[x]);
        for (const y of ys) {
            const value = used[x][y];
            const count = result[value];

            result[value] = count + 1;
        }
    }
    return result;
}

const renderBoard = (state: AppState, dispatch) => {
    const svgContainer = document.createElement("div")

    if (state.boardState.gameState === GameStates.STARTED) {
        const counts = document.createElement("div")
        svgContainer.append(counts)

        const countValues = getCounts(state.boardState.used);
        const blackCount = document.createElement("span")
        blackCount.setAttribute("style", "padding: 10px;")
        blackCount.innerText = `BLACK SCORE: is ${countValues[Sides.BLACK]}`

        const whiteCount = document.createElement("span")
        whiteCount.setAttribute("style", "padding: 10px;")
        whiteCount.innerText = `WHITE SCORE: is ${countValues[Sides.WHITE]}`

        counts.append(blackCount)
        counts.append(whiteCount)
        counts.setAttribute("style", "text-align: center;")

        const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
        svg.setAttribute("viewBox", "0 0 171 191")
        svg.setAttribute("xmlns", "http://www.w3.org/2000/svg")
        svg.setAttribute("width", "800")
        svg.setAttribute("height", "1000")
        svg.setAttribute("style", ["padding: 10px", "display: block", "margin: auto"].join(";"))

        const { used, playerTurn, availableMoves, showMoves, edge } = state.boardState;

        const currentPlayerTurn = playerTurn === state.side;
        if (currentPlayerTurn) {
            const active = document.createElementNS("http://www.w3.org/2000/svg", "rect")
            active.setAttribute("width", "171")
            active.setAttribute("height", "171")
            active.setAttribute("x", "0")
            active.setAttribute("y", "0")
            active.setAttribute("fill", "yellow")

            svg.appendChild(active)
        }
        range(0, 8).forEach((x) => {
            range(0, 8).forEach((y) => {
                const square = document.createElementNS("http://www.w3.org/2000/svg", "rect")
                square.setAttribute("width", "20")
                square.setAttribute("height", "20")
                square.setAttribute("x", `${(x * 21) + 2}`)
                square.setAttribute("y", `${(y * 21) + 2}`)

                if (r.find(({ X, Y }) => X === x && Y === y)(availableMoves) && currentPlayerTurn && showMoves) {
                    square.setAttribute("fill", "grey")
                } else {
                    square.setAttribute("fill", "green")
                }

                svg.appendChild(square)

                ifClaimedOrElse({
                    data: {
                        claimed: used,
                        coordinate: { x, y }
                    },
                    ifClaimed: (sideColor: string) => {
                        const piece = document.createElementNS("http://www.w3.org/2000/svg", "circle")
                        piece.setAttribute("r", "8")
                        piece.setAttribute("cx", `${(x * 21) + 2 + 8 + 2}`)
                        piece.setAttribute("cy", `${(y * 21) + 2 + 8 + 2}`)
                        piece.setAttribute("fill", sideColor)

                        svg.appendChild(piece)
                    },
                    orElse: () => {
                        square.onclick = () => {
                            socket.send(JSON.stringify({ X: x, Y: y }))
                        }
                    }
                })
            })
        })

        const { showColor, hideColor } = getShowHideColors({ showMoves })

        const show = document.createElementNS("http://www.w3.org/2000/svg", "rect")
        const showText = document.createElementNS("http://www.w3.org/2000/svg", "text")
        showText.textContent = "SHOW";
        showText.setAttribute("x", "21")
        showText.setAttribute("y", "190")
        showText.setAttribute("width", "20")
        showText.setAttribute("height", "10")
        showText.onclick = () => dispatch(getShow())

        show.setAttribute("x", "21")
        show.setAttribute("y", "170")
        show.setAttribute("width", "20")
        show.setAttribute("height", "10")
        show.setAttribute("fill", showColor)
        show.onclick = () => dispatch(getShow())

        const hide = document.createElementNS("http://www.w3.org/2000/svg", "rect")
        const hideText = document.createElementNS("http://www.w3.org/2000/svg", "text")
        hideText.textContent = "HIDE";
        hideText.setAttribute("x", "126")
        hideText.setAttribute("y", "190")
        hideText.setAttribute("width", "20")
        hideText.setAttribute("height", "10")
        hideText.onclick = () => dispatch(getHide())

        hide.setAttribute("x", "126")
        hide.setAttribute("y", "170")
        hide.setAttribute("width", "20")
        hide.setAttribute("height", "10")
        hide.setAttribute("fill", hideColor)
        hide.onclick = () => dispatch(getHide())

        svg.appendChild(show)
        svg.appendChild(showText)
        svg.appendChild(hide)
        svg.appendChild(hideText)

        svgContainer.appendChild(svg)
    }

    return svgContainer;
}

const render = (rootElement: HTMLElement) => (state: AppState, dispatch) => {
    const child = rootElement.firstChild;
    if (child) {
        rootElement.removeChild(rootElement.firstChild)
    }

    const headerElement = document.createElement("div")
    const sideElement = document.createElement("div")
    const turnElement = document.createElement("div")

    headerElement.append(sideElement)
    headerElement.append(turnElement)

    sideElement.innerText = `YOUR SIDE: ${state.side}`;
    let turnText = ""
    if (state.boardState.gameState === GameStates.STARTED) {
        if (state.boardState.playerTurn === state.side) {
            turnText = "Your Turn"
        } else {
            turnText = "Other Player's Turn"
        }
        turnElement.innerText = turnText;
    }

    const messageElement = document.createElement("pre")
    messageElement.classList.add("message")
    messageElement.innerText = JSON.stringify(state, null, 4)

    const documentContainer = document.createElement("div")
    documentContainer.append(headerElement)
    documentContainer.append(renderBoard(state, dispatch))
    documentContainer.append(messageElement)

    rootElement.appendChild(documentContainer)
}

const store = getStore(reducer)
store.subscribe(render(document.querySelector(".app")))

socket.addEventListener('open', (event) => {
})

socket.addEventListener('message', (event) => {
    console.log(`Got a message ${event.data}`)
    withAction(event.data, store.dispatch)
})