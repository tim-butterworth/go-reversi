interface Action {
    actionType: string;
}

type Dispatch<A extends Action> = (a: A) => void;
type Subscriber<S, A extends Action> = (state: S, dispatch: Dispatch<A>) => void;

type InitialAction = {
    actionType: "INITIALIZE"
}
type Reducer<S, A extends Action> = (args: { action: A | InitialAction, state?: S }) => S;

interface Store<S, A extends Action> {
    subscribe: (s: Subscriber<S, A>) => void;
    dispatch: Dispatch<A>;
}

const getStore = <S, A extends Action>(reducers: Reducer<S, A>): Store<S, A> => {
    let state = reducers({ action: { actionType: "INITIALIZE" } });

    const subscribers = [];
    const subscribe = <S>(subscriber: Subscriber<S, A>) => {
        subscribers.push(subscriber);
    }

    const dispatch = (action) => {
        const updatedState = reducers({ state, action });

        if (updatedState !== state) {
            state = updatedState
            subscribers.forEach((s) => s(state, dispatch))
        }
    }
    return {
        dispatch,
        subscribe
    }
}

export {
    getStore
}
