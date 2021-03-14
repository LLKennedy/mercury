import { EnumType } from "typescript";
import { EnumMap } from "./Enums";

/** Converts an object  */
export type Parser<T> = (res: any) => Promise<T>;

/** These are all allowed basic types returned by the typeof accessor */
export type TypeStrings = "string" | "number" | "bigint" | "boolean" | "symbol" | "undefined" | "object" | "function";

/** Converts data structured as "any" to explicitly Object type, only supports string parsing and object assertion at present. */
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

/** Runs the provided set function with the acquired value if the object has the specified property and the value is not null. Optionally throws an error if typeof returns an unsupported type */
export async function SetIfNotNull<T>(obj: Object, prop: string, set: (val: any) => Promise<T | undefined>, validTypes: TypeStrings[] = ["string", "object", "boolean", "number", "undefined", "bigint", "function", "symbol"]) {
	if (obj.hasOwnProperty(prop)) {
		let resNum = obj[prop]
		if (resNum !== null) {
			let val = obj[prop];
			if (!validTypes.includes(typeof val)) {
				throw new Error(`invalid type for property ${prop}, exptected one of ${validTypes} but found ${typeof val}`);
			}
			return await set(val);
		}
	}
	return undefined;
}

/** Parsing functions for all canonical gRPC JSON types
 * 
 * Definitions and behaviour according to:
 * https://developers.google.com/protocol-buffers/docs/proto3#json
 */
export class Parse {
	/** Parse a message object. This is just the same logic as using parser directly. */
	public static async Message<T>(obj: Object, prop: string, parser: Parser<T>): Promise<T | undefined> {
		return SetIfNotNull(obj, prop, async raw => {
			if (typeof raw !== "object") {
				throw new Error(`message types must be objects, found ${typeof raw} instead`)
			}
			return parser(raw);
		}, ["object"]);

	}
	/** Parse an enum which could be either strings or numbers. This is NOT fully type safe, if EnumMap is not the Object.keys of the actual enum T, bad things will happen */
	public static async Enum<T extends EnumType>(obj: Object, prop: string, map: EnumMap): Promise<T | undefined> {
		return SetIfNotNull(obj, prop, async raw => {
			switch (typeof raw) {
				case "number":
					return raw as unknown as T;
				case "string":
					return map[raw] as unknown as T;
				default:
					throw new Error(`enum types must be strings or numbers, found ${typeof raw} instead`)
			}
		}, ["string", "number"]);
	}
	public static async Map<K, V>(obj: Object, prop: string, keyParse: (key: string) => Promise<K>, valParse: (val: any) => Promise<V>): Promise<ReadonlyMap<K, V> | undefined> {
		return SetIfNotNull(obj, prop, async raw => {
			if (typeof raw !== "object") {
				throw new Error(`map types must be objects, found ${typeof raw} instead`)
			}
			throw new Error("unimplemented - we need to write all keyParse and valParse functions")
		}, ["object"]);
	}
}