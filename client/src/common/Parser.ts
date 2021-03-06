/** Converts an object  */
export type Parser<T> = (res: any) => Promise<T>;
