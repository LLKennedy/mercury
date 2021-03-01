import { ProtoJSONCompatible } from "../common";
import { EstablishedWebsocket } from "./websocket";

/** DualStream is a dual streaming RPC */
export class DualStream<ReqT extends ProtoJSONCompatible, ResT = any> {
	private ws: EstablishedWebsocket<ReqT, ResT>;
	constructor(ws: EstablishedWebsocket<ReqT, ResT>) {
		this.ws = ws;
	}
	/** Wait until a message is successfully sent to the server. */
	public Send(request: ReqT): Promise<void> {
		return this.ws.Send(request);
	}
	/** Wait until a message is received from the server. */
	public Recv(): Promise<ResT> {
		return this.ws.Recv();
	}
	/** Close the sending direction of communications, any Send calls after this will throw an Error without writing to the websocket. */
	public CloseSend(): Promise<void> {
		return this.ws.CloseSend();
	}
}

/** DualStream is a dual streaming RPC */
export interface IDualStream<ReqT extends ProtoJSONCompatible, ResT = any> {
	/** Wait until a message is successfully sent to the server. */
	Send(request: ReqT): Promise<void>;
	/** Wait until a message is received from the server. */
	Recv(): Promise<ResT>;
	/** Close the sending direction of communications, any Send calls after this will throw an Error without writing to the websocket. */
	CloseSend(): Promise<void>;
}