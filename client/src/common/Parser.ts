import { EnumType } from "typescript";
import { EnumMap } from "./Enums";
import { base64 } from "rfc4648"
import { google } from ".";

/** Converts an object  */
export type Parser<T> = (res: any) => Promise<T>;
export type RepeatedParser<T> = (res: any[]) => Promise<T[]>;

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
export async function ParseIfNotNull<T>(obj: Object, prop: string, set: (val: any) => Promise<T | undefined>, validTypes: TypeStrings[] = ["string", "object", "boolean", "number", "undefined", "bigint", "function", "symbol"]) {
	if (obj.hasOwnProperty(prop)) {
		let foundProp = obj[prop];
		if (foundProp !== null && foundProp !== undefined) {
			if (!validTypes.includes(typeof foundProp)) {
				throw new Error(`invalid type for property ${prop}, exptected one of ${validTypes} but found ${typeof foundProp}`);
			}
			return await set(foundProp);
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
		return ParseIfNotNull(obj, prop, PrimitiveParse.Message<T>(parser), ["object"]);
	}
	/** Parse an enum which could be either strings or numbers. This is NOT fully type safe, if EnumMap is not the Object.keys of the actual enum T, bad things will happen */
	public static async Enum<T extends EnumType>(obj: Object, prop: string, map: EnumMap): Promise<T | undefined> {
		return ParseIfNotNull(obj, prop, PrimitiveParse.Enum<T>(map), ["string", "number"]);
	}
	/** Parse a map, providing individual parsers for key and value instances */
	public static async Map<K, V>(obj: Object, prop: string, keyParse: (key: string) => Promise<K>, valParse: (val: any) => Promise<V>): Promise<ReadonlyMap<K, V> | undefined> {
		return ParseIfNotNull(obj, prop, PrimitiveParse.Map<K, V>(keyParse, valParse), ["object"]);
	}
	/** Parse an array */
	public static async Repeated<T>(obj: Object, prop: string, parser: Parser<T>): Promise<T[] | undefined> {
		return ParseIfNotNull(obj, prop, PrimitiveParse.Repeated<T>(parser), ["object"]);
	}
	/** Parse a boolean */
	public static async Bool(obj: Object, prop: string): Promise<boolean | undefined> {
		return ParseIfNotNull(obj, prop, PrimitiveParse.Bool(), ["boolean"]);
	}
	/** Parse a string */
	public static async String(obj: Object, prop: string): Promise<string | undefined> {
		return ParseIfNotNull(obj, prop, PrimitiveParse.String(), ["string"]);
	}
	/** Parse bytes */
	public static async Bytes(obj: Object, prop: string): Promise<Uint8Array | undefined> {
		return ParseIfNotNull(obj, prop, PrimitiveParse.Bytes(), ["string"]);
	}
	/** Parse a number */
	public static async Number(obj: Object, prop: string): Promise<number | undefined> {
		return ParseIfNotNull(obj, prop, PrimitiveParse.Number(), ["string", "number"]);
	}
	/** Parse a google Any */
	public static async Any(obj: Object, prop: string): Promise<any | undefined> {
		return ParseIfNotNull(obj, prop, google.Any.Parse)
	}
	/** Parse a google Timestamp */
	public static async Timestamp(obj: Object, prop: string): Promise<Date | undefined> {
		return ParseIfNotNull(obj, prop, google.Timestamp.Parse)
	}
	/** Parse a google Duration */
	public static async Duration(obj: Object, prop: string): Promise<number | undefined> {
		return ParseIfNotNull(obj, prop, google.Duration.Parse)
	}
	/** Parse a google Struct */
	public static async Struct(obj: Object, prop: string): Promise<any | undefined> {
		return ParseIfNotNull(obj, prop, google.Struct.Parse)
	}
	/** Parse a google Wrapper */
	public static async Wrapper(obj: Object, prop: string): Promise<any | undefined> {
		return ParseIfNotNull(obj, prop, google.Wrapper.Parse)
	}
	/** Parse a google FieldMask */
	public static async FieldMask(obj: Object, prop: string): Promise<any | undefined> {
		return ParseIfNotNull(obj, prop, google.FieldMask.Parse)
	}
	/** Parse a google ListValue */
	public static async ListValue(obj: Object, prop: string): Promise<any[] | undefined> {
		return ParseIfNotNull(obj, prop, google.ListValue.Parse)
	}
	/** Parse a google Value */
	public static async Value(obj: Object, prop: string): Promise<any | undefined> {
		return ParseIfNotNull(obj, prop, google.Value.Parse)
	}
	/** Parse a google NullValue */
	public static async NullValue(obj: Object, prop: string): Promise<null> {
		// I don't know why anyone would ever use this type
		return null
	}
	/** Parse a google Empty */
	public static async Empty(obj: Object, prop: string): Promise<any | undefined> {
		return ParseIfNotNull(obj, prop, google.Any.Parse)
	}
}

export class PrimitiveParse {
	public static Message<T>(parser: Parser<T>): Parser<T> {
		return async raw => {
			if (typeof raw !== "object") {
				throw new Error(`message types must be objects, found ${typeof raw} instead`)
			}
			return parser(raw);
		}
	}
	public static Enum<T extends EnumType>(map: EnumMap): Parser<T> {
		return async raw => {
			if (typeof raw === "string" && raw === "") {
				// Empty string is the zero value
				raw = 0;
			}
			switch (typeof raw) {
				case "number":
					let mappedStr = map[raw];
					if (mappedStr === undefined) {
						throw new Error(`undefined enum value: ${raw}`);
					}
					return raw as unknown as T;
				case "string":
					let mappedNum = map[raw] as unknown as T;
					if (mappedNum === undefined) {
						throw new Error(`undefined enum value: ${raw}`)
					}
					return mappedNum;
				default:
					throw new Error(`enum types must be strings or numbers, found ${typeof raw} instead`)
			}
		}
	}
	public static Map<K, V>(keyParse: (key: string) => Promise<K>, valParse: (val: any) => Promise<V>): Parser<ReadonlyMap<K, V>> {
		return async raw => {
			if (typeof raw !== "object") {
				throw new Error(`map types must be objects, found ${typeof raw} instead`)
			}
			throw new Error("unimplemented - we need to write all keyParse and valParse functions")
		}
	}
	public static Repeated<T>(parser: Parser<T>): RepeatedParser<T> {
		return async raw => {
			if (!(raw instanceof Array)) {
				throw new Error(`array type expected, found ${raw} instead`)
			}
			throw new Error("unimplemented")
		}
	}
	public static Bool(): Parser<boolean> {
		return async raw => {
			if (typeof raw !== "boolean") {
				throw new Error(`boolean type expected, found ${typeof raw} instead`);
			}
			return raw
		}
	}
	public static String(): Parser<string> {
		return async raw => {
			if (typeof raw !== "string") {
				throw new Error(`string type expected, found ${typeof raw} instead`);
			}
			return raw
		}
	}
	public static Bytes(): Parser<Uint8Array> {
		return async raw => {
			if (typeof raw !== "string") {
				throw new Error(`string type expected, found ${typeof raw} instead`);
			}
			return base64.parse(raw);
		}
	}
	/** int32, fixed32, uint32, int64, fixed64, uint64, float, double - all work on identical logic other than range checking */
	public static Number(rangeCheck?: (num: number) => boolean, allowSpecial: boolean = false): Parser<number> {
		return async raw => {
			if (typeof raw !== "number" && typeof raw !== "string") {
				throw new Error(`string or number type expected`)
			}
			let parsed: number;
			if (typeof raw === "number") {
				parsed = raw;
			} else {
				parsed = Number(raw);
			}
			if (isNaN(parsed) || parsed === Infinity || parsed === -Infinity) {
				if (allowSpecial) {
					return parsed;
				}
				throw new Error("received special number (NaN or Infinity) when not allowed");
			}
			if (rangeCheck && !rangeCheck(parsed)) {
				throw new Error("parsed number failed range check")
			}
			return parsed;
		}
	}
}