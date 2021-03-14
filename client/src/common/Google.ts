/**
 * This file is for special types exported by google protobufs, such as Any, Empty, Timestamp and Duration.
 * They get special treatment by the canonical JSON marshalling rules, so we give them special treatment here too.
 */

import { ProtoJSONCompatible } from "./ProtoJSONCompatible";

export class Any implements ProtoJSONCompatible {
	public value?: any;
	public ToProtoJSON(): Object {
		throw new Error("unimplemented");
	}
	public static async Parse(data: any): Promise<any> {
		throw new Error("unimplemented");
	}
}

export class Timestamp implements ProtoJSONCompatible {
	public timestamp?: Date;
	public ToProtoJSON(): Object {
		throw new Error("unimplemented");
	}
	public static async Parse(data: any): Promise<Date> {
		throw new Error("unimplemented");
	}
}

export class Duration implements ProtoJSONCompatible {
	public durationSeconds?: number;
	public ToProtoJSON(): Object {
		throw new Error("unimplemented");
	}
	public static async Parse(data: any): Promise<number> {
		throw new Error("unimplemented");
	}
}

export class Struct implements ProtoJSONCompatible {
	public data?: any;
	public ToProtoJSON(): Object {
		throw new Error("unimplemented");
	}
	public static async Parse(data: any): Promise<any> {
		throw new Error("unimplemented");
	}
}

export class Wrapper implements ProtoJSONCompatible {
	public ToProtoJSON(): Object {
		throw new Error("unimplemented");
	}
	public static async Parse(data: any): Promise<any> {
		throw new Error("unimplemented");
	}
}

export class FieldMask implements ProtoJSONCompatible {
	public ToProtoJSON(): Object {
		throw new Error("unimplemented");
	}
	public static async Parse(data: any): Promise<any> {
		throw new Error("unimplemented");
	}
}

export class ListValue implements ProtoJSONCompatible {
	public list?: any[];
	public ToProtoJSON(): Object {
		throw new Error("unimplemented");
	}
	public static async Parse(data: any): Promise<any[]> {
		throw new Error("unimplemented");
	}
}

export class Value implements ProtoJSONCompatible {
	public value?: any;
	public ToProtoJSON(): Object {
		throw new Error("unimplemented");
	}
	public static async Parse(data: any): Promise<any> {
		throw new Error("unimplemented");
	}
}

export class NullValue implements ProtoJSONCompatible {
	public ToProtoJSON(): Object {
		throw new Error("unimplemented");
	}
	public static async Parse(data: any): Promise<null> {
		return null;
	}
}

export class Empty implements ProtoJSONCompatible {
	public ToProtoJSON(): Object {
		throw new Error("unimplemented");
	}
	public static async Parse(data: any): Promise<{}> {
		return {}
	}
}