import { ProtoJSONCompatible } from "@llkennedy/protoc-gen-tsjson";
import { StreamBase } from "./streambase";
import { EstablishedWebsocket } from "./websocket";

/** DualStream is a dual streaming RPC */
export class DualStream<ReqT extends ProtoJSONCompatible, ResT = any> extends StreamBase<ReqT, ResT> {
	constructor(ws: EstablishedWebsocket<ReqT, ResT>) {
		super(ws);
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
export interface IDualStream<ReqT extends ProtoJSONCompatible, ResT = any> extends DualStream<ReqT, ResT> { }