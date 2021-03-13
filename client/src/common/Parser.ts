/** Converts an object  */
export type Parser<T> = (res: any) => Promise<T>;

export type TypeStrings = "string" | "number" | "bigint" | "boolean" | "symbol" | "undefined" | "object" | "function";

export function AnyToObject(res: any): Object {
	switch (typeof res) {
		case "string":
			return JSON.parse(res);
		case "object":
			return res;
		default:
			throw new Error("only string and object parsing are supported")
	}
}

export async function SetIfNotNull(obj: Object, prop: string, set: (val: any) => Promise<void>, validTypes: TypeStrings[] = ["string", "object", "boolean", "number", "undefined", "bigint", "function", "symbol"]) {
	if (obj.hasOwnProperty(prop)) {
		let resNum = obj[prop]
		if (resNum !== null) {
			let val = obj[prop];
			if (!validTypes.includes(typeof val)) {
				throw new Error(`invalid type for property ${prop}, exptected one of ${validTypes} but found ${typeof val}`);
			}
			await set(val);
		}
	}
}