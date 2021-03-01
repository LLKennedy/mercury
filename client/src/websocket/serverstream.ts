import { ProtoJSONCompatible } from "common";
import { EstablishedWebsocket } from "websocket";

/** ServerStream is a server-side streaming RPC */
export class ServerStream<ReqT extends ProtoJSONCompatible, ResT = any> {
	private ws: EstablishedWebsocket<ReqT, ResT>;
	private request: ReqT;
	private initialised: boolean = false;
	constructor(ws: EstablishedWebsocket<ReqT, ResT>, request: ReqT) {
		this.ws = ws;
		this.request = request;
	}
	/** Sends the first request.
	 * 
	 * Can only be called once.
	 * 
	 * This will normally be called by the Client method which establishes the connection, you only need to use this method
	 * if you are creating instances of this class directly for some reason.
	 * 
	 * Calling this a second time will always throw an error without doing anything else.
	 */
	public async init(): Promise<IServerStream<ResT>> {
		if (this.initialised) {
			throw new Error("cannot initialise ServerStream twice");
		}
		this.initialised = true;
		await this.ws.Send(this.request);
		return this;
	}
	/** Wait until a message is received from the server. */
	public Recv(): Promise<ResT> {
		return this.ws.Recv();
	}
}

/** ServerStream is a server-side streaming RPC */
export interface IServerStream<ResT = any> {
	/** Wait until a message is received from the server. */
	Recv(): Promise<ResT>;
}