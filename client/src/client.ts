import axios from "axios";

export class Client {
	private basePath: string;
	private secure: boolean;
	constructor(basePath: string = "localhost/api", secure: boolean = true) {
		this.basePath = basePath;
		this.secure = secure;
	}
	// TODO: this request probably shouldn't be serialisable, that interface should be for websockets
	public async SendUnary<ResT = any>(endpoint: string, request: Serialisable, parseResponse: ((res: Object) => ResT) | ((res: Object) => Promise<ResT>)): Promise<ResT> {
		let res = await axios.post<Object>(this.BuildURL(endpoint), request.Serialise());
		return parseResponse(res.data);
	}
	public BuildURL(endpoint: string, websocket: boolean = false): string {
		// First get the scheme
		let scheme: string;
		switch (websocket) {
			case true:
				switch (this.secure) {
					case true:
						scheme = "wss";
						break;
					case false:
						scheme = "ws";
						break;
				}
				break;
			case false:
				switch (this.secure) {
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

export interface Serialisable {
	Serialise(): Object
}

export default Client;