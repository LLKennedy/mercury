import axios, { AxiosInstance as AI } from "axios";
import { ClientStream, DualStream, MercuryWebSocket, IClientStream, IDualStream, IServerStream, ServerStream } from "../websocket";
import { ProtoJSONCompatible, Parser } from "@llkennedy/protoc-gen-tsjson";

export interface AxiosInstance extends AI { };
export const DefaultAxios = axios;

/** Client is an RPC client proxied over HTTP and websockets. It is recommended to wrap this in service-specific RPC definitions, 
 * rather than relying on end-users to use the type parameters correctly. */
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
	protected async SendUnary<ReqT extends ProtoJSONCompatible, ResT = any>(endpoint: string, method: HTTPMethod, request: ReqT, parseResponse: Parser<ResT>): Promise<ResT> {
		let message = request.ToProtoJSON();
		let url = this.BuildURL(endpoint, false);
		switch (method) {
			case HTTPMethod.CONNECT:
				throw new Error("CONNECT not implemented");
			case HTTPMethod.TRACE:
				throw new Error("TRACE not implemented");
		}
		let res = await this.axiosClient.request<Object>({
			url: url,
			method: method,
			data: message
		})
		return parseResponse(res.data);
	}
	protected async StartClientStream<ReqT extends ProtoJSONCompatible, ResT = any>(endpoint: string, parseResponse: Parser<ResT>): Promise<ClientStream<ReqT, ResT>> {
		let url = this.BuildURL(endpoint, true);
		let ws = new MercuryWebSocket(url, parseResponse);
		// Establish the connection, set up event listeners, etc.
		await ws.init();
		return new ClientStream(ws);
	}
	protected async StartServerStream<ReqT extends ProtoJSONCompatible, ResT = any>(endpoint: string, request: ReqT, parseResponse: Parser<ResT>): Promise<ServerStream<ReqT, ResT>> {
		let url = this.BuildURL(endpoint, true);
		let ws = new MercuryWebSocket(url, parseResponse);
		// Establish the connection, set up event listeners, etc.
		await ws.init();
		let ss = new ServerStream<ReqT, ResT>(ws, request);
		return ss.init();
	}
	protected async StartDualStream<ReqT extends ProtoJSONCompatible, ResT = any>(endpoint: string, parseResponse: Parser<ResT>): Promise<DualStream<ReqT, ResT>> {
		let url = this.BuildURL(endpoint, true);
		let ws = new MercuryWebSocket(url, parseResponse);
		// Establish the connection, set up event listeners, etc.
		await ws.init();
		return new DualStream(ws);
	}
	protected BuildURL(endpoint: string, websocket: boolean = false): string {
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

export enum HTTPMethod {
	GET = "GET",
	HEAD = "HEAD",
	POST = "POST",
	PUT = "PUT",
	DELETE = "DELETE",
	CONNECT = "CONNECT",
	OPTIONS = "OPTIONS",
	TRACE = "TRACE",
	PATCH = "PATCH"
}

export default Client;