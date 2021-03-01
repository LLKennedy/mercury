/** Converts an object  */
export type Parser<T> = (res: Object) => Promise<T>;
