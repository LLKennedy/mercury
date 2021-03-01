import { ProtoJSONCompatible } from "common";
import { EstablishedWebsocket } from "websocket";

/** ClientStream is a client-side streaming RPC */
export class ClientStream<ReqT extends ProtoJSONCompatible, ResT = any> {
	private ws: EstablishedWebsocket<ReqT, ResT>;
	private closed: boolean = false;
	constructor(ws: EstablishedWebsocket<ReqT, ResT>) {
		this.ws = ws;
	}
	/** Wait until a message is successfully sent to the server. */
	public Send(request: ReqT): Promise<void> {
		if (this.closed) {
			throw new Error("Cannot use closed ClientStream");
		}
		return this.ws.Send(request);
	}
	/** Close the sending stream and wait for a single server-side response. After this the stream cannot be used. */
	public async CloseAndRecv(): Promise<ResT> {
		if (this.closed) {
			throw new Error("Cannot use closed ClientStream");
		}
		this.closed = true;
		await this.ws.CloseSend();
		return this.ws.Recv();
	}
}

export interface IClientStream<ReqT extends ProtoJSONCompatible, ResT = any> {
	/** Wait until a message is successfully sent to the server. */
	Send(request: ReqT): Promise<void>;
	/** Close the sending stream and wait for a single server-side response. After this the stream cannot be used. */
	CloseAndRecv(): Promise<ResT>
}