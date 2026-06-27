// A simple signal implementation.
// For example, we can use a Signal to allow waiting for a decision
// from the user in a Modal and then continue logic depending on the response.

export type SignalFireFunc<T> = (value: T) => void;

export interface Signal<T> {
	promise: Promise<T>;
	fire: SignalFireFunc<T>;
}

/**
 * Create a new signal.
 * Returns a new promise that we can await.
 * The `fire` function resolves the promise.
 */
export function createSignal<T>(): Signal<T> {
	let resolve: SignalFireFunc<T>;
	const promise = new Promise<T>((r) => (resolve = r));
	return {
		promise,
		fire: (data: T) => resolve(data),
	};
}
