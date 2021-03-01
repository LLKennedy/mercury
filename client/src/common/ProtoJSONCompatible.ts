/** Many messages are simpler to build and manage using native types that aren't 100% identical to what is expected by the canonical JSON representation of those messages.
 *
 * To deal with this, you are encouraged to use classes that hold those native values, but implement a "Serialise" function that converts them to the protojson format.
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
