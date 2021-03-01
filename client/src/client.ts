import axios, { AxiosInstance } from "axios";
import { ClientStream, DualStream, EstablishedWebsocket, HTTPgRPCWebSocket, IClientStream, IDualStream, IServerStream, ServerStream } from "websocket";

export type Parser<T> = ((res: Object) => T) | ((res: Object) => Promise<T>);

export class Client {
	private axiosClient: AxiosInstance;
	private basePath: string;
	private useTLS: boolean;
	constructor(basePath: string = "localhost/api", useTLS: boolean = true, client: AxiosInstance = axios) {
		this.basePath = basePath;
		this.useTLS = useTLS;
		this.axiosClient = client;
	}
	/**
	 * @param {string} endpoint The API endpoint to send the message to
	 * @param {ProtoJSONCompatible} request The message to convert to protojson then send
	 */
	public async SendUnary<ReqT extends ProtoJSONCompatible, ResT = any>(endpoint: string, request: ReqT, parseResponse: Parser<ResT>): Promise<ResT> {
		let message = request.ToProtoJSON();
		let url = this.BuildURL(endpoint, false);
		let res = await this.axiosClient.post<Object>(url, message);
		return parseResponse(res.data);
	}
	public async StartClientStream<ReqT extends ProtoJSONCompatible, ResT = any>(endpoint: string, parseResponse: Parser<ResT>): Promise<IClientStream<ReqT, ResT>> {
		let url = this.BuildURL(endpoint, true);
		let ws = new HTTPgRPCWebSocket(url, parseResponse);
		// Establish the connection, set up event listeners, etc.
		await ws.init();
		return new ClientStream(ws);
	}
	public async StartServerStream<ReqT extends ProtoJSONCompatible, ResT = any>(endpoint: string, request: ReqT, parseResponse: Parser<ResT>): Promise<IServerStream<ResT>> {
		let url = this.BuildURL(endpoint, true);
		let ws = new HTTPgRPCWebSocket(url, parseResponse);
		// Establish the connection, set up event listeners, etc.
		await ws.init();
		let ss = new ServerStream(ws, request);
		return ss.init();
	}
	public async StartDualStream<ReqT extends ProtoJSONCompatible, ResT = any>(endpoint: string, parseResponse: Parser<ResT>): Promise<IDualStream<ReqT, ResT>> {
		let url = this.BuildURL(endpoint, true);
		let ws = new HTTPgRPCWebSocket(url, parseResponse);
		// Establish the connection, set up event listeners, etc.
		await ws.init();
		return new DualStream(ws);
	}
	public BuildURL(endpoint: string, websocket: boolean = false): string {
		// First get the scheme
		let scheme: string;
		switch (websocket) {
			case true:
				switch (this.useTLS) {
					case true:
						scheme = "wss";
						break;
					case false:
						scheme = "ws";
						break;
				}
				break;
			case false:
				switch (this.useTLS) {
					case true:
						scheme = "https";
						break;
					case false:
						scheme = "http";
						break;
				}
				break;
		}
		return `${scheme}://${this.basePath}/${endpoint}`;
	}
}

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
	ToProtoJSON(): Object
}

export default Client;