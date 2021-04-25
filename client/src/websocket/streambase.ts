import { ProtoJSONCompatible } from "@llkennedy/protoc-gen-tsjson";
import { EstablishedWebsocket } from "./websocket";

/** StreamBase is the base class for stream handler implementations. Not intended for direct use. */
export class StreamBase<ReqT extends ProtoJSONCompatible, ResT = any> {
	protected ws: EstablishedWebsocket<ReqT, ResT>;
	constructor(ws: EstablishedWebsocket<ReqT, ResT>) {
		this.ws = ws;
	}
	/** CloseEarly terminates the connection to the server, causing errors upstream. Use to abort in-progress calls and clean up the websocket. */
	public async CloseEarly(code?: number, reason?: string): Promise<void> {
		await this.ws.Close(code, reason);
	}
}