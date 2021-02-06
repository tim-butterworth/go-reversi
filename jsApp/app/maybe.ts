enum MaybeType {
    NONE = "NONE",
    SOME = "SOME"
}
type None = {
    maybeType: MaybeType.NONE
}
type Some<T> = {
    maybeType: MaybeType.SOME,
    data: T,
}
type Maybe<T> = Some<T> | None;

const some = <T>(t: T): Some<T> => ({
    maybeType: MaybeType.SOME,
    data: t
})
const none = (): None => ({
    maybeType: MaybeType.NONE
})

type Effect<T> = (t: T) => void;
const ifPresent = <T>(m: Maybe<T>, effect: Effect<T>): void => {
    if (m.maybeType === MaybeType.SOME) {
        effect(m.data)
    }
}

export {
    Maybe,
    some,
    none,
    ifPresent
}
