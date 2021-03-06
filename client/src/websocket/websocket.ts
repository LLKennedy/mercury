import { ProtoJSONCompatible, Parser } from "../common";

const EOFMessage = "EOF";

export class HTTPgRPCWebSocket<ReqT extends ProtoJSONCompatible, ResT = any> {
	private parser: Parser<ResT>;
	private url: string;
	private initialised: boolean = false;
	private conn?: WebSocket = undefined;
	private responseBuffer: string[] = [];
	constructor(url: string, parser: Parser<ResT>) {
		this.parser = parser;
		this.url = url;
	}
	/** Attempts to establish a websocket connection, then set up event listeners to handle bi-directional comms.
	 * 
	 * Can only be called once.
	 * 
	 * This will normally be called by the Client method which establishes the connection, you only need to use this method
	 * if you are creating instances of this class directly for some reason.
	 * 
	 * Calling this a second time will always throw an error without doing anything else.
	 */
	public async init(): Promise<EstablishedWebsocket<ReqT, ResT>> {
		if (this.initialised) {
			throw new Error("cannot initialise HTTPgRPCWebSocket twice");
		}
		this.initialised = true;
		this.conn = new WebSocket(this.url);
		return this;
	}
	/** Wait until a message is successfully sent to the server. */
	public async Send(request: ReqT): Promise<void> {
		let message = request.ToProtoJSON();
		let messageString = JSON.stringify(message);
		this.conn?.send(messageString);
	}
	/** Wait until a message is received from the server. */
	public async Recv(): Promise<ResT> {
		let next = this.responseBuffer.shift();
		if (next === undefined) {
			// FIXME this isn't how buffers work, redo this whole section
			throw new Error("unimplemented")
		}
		return this.parser(next);
	}
	/** Close the sending direction of communications, any Send calls after this will throw an Error without writing to the websocket. */
	public async CloseSend(): Promise<void> {
		this.conn?.send(EOFMessage);
	}
	/** Close terminates the websocket connection early, resulting in errors at the server's end. Any other calls after this point will throw an Error. */
	public async Close(code?: number, reason?: string): Promise<void> {
		this.conn?.close(code, reason);
	}
}

/** A websocket connection open and ready to handle gRPC message transfer in both directions.
 * 
 * Wrap this in one of the specialised gRPC streaming patterns if you wish to simplify the 
 * API and not have to worry about when to call CloseSend, etc.
 */
export interface EstablishedWebsocket<ReqT extends ProtoJSONCompatible, ResT = any> {
	/** Wait until a message is successfully sent to the server. */
	Send(request: ReqT): Promise<void>;
	/** Wait until a message is received from the server. */
	Recv(): Promise<ResT>;
	/** Close the sending direction of communications, any Send calls after this will throw an Error without writing to the websocket. */
	CloseSend(): Promise<void>;
	/** Close terminates the websocket connection early, resulting in errors at the server's end. Any other calls after this point will throw an Error. */
	Close(code?: number, reason?: string): Promise<void>;
}