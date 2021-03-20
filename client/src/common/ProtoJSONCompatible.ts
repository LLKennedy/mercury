/** Many messages are simpler to build and manage using native types that aren't 100% identical to what is expected by the canonical JSON representation of those messages.
 *
 * To deal with this, you are encouraged to use classes that hold those native values, but implement a "ToProtoJSON" function that converts them to the protojson format.
 *
 * For messages sent on unary RPCs, these objects will be sent as-is on the provided axios client in HTTP requests.
 *
 * For messages sent on streamed RPCs, these objects will be passed through JSON.stringify and sent on the websocket send channel.
 */
export interface ProtoJSONCompatible {
	/** Convert native fields to canonical protojson format
	 *
	 * e.g. 64-bit numbers as strings, bytes as base64, oneofs as specific instance fields
	 * */
	ToProtoJSON(): Object;
}

/**
 * AssignFields is a shorthand for Object.Assign with a partial message. This acts as an automatically typed
 * extension to the constructor, so you can call e.g. AssignFields(new MyMessageType(), {validField: "value"})
 * with no additional codegen required.
 * @param base 
 * @param fields 
 */
export function AssignFields<T extends ProtoJSONCompatible>(base: T, fields: Partial<T>): T {
	return Object.assign<T, Partial<T>>(base, fields);
}