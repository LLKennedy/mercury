import { ProtoJSONCompatible, Parser, IMutex, Mutex } from "../common";
import * as uuid from "uuid";

const EOFMessage = "EOF";

/** A logger that wraps the standard console log functions */
export interface Logger {
	log(msg: string);
	warn(msg: string);
	error(msg: string);
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

/** A websocket connection open and ready to handle gRPC message transfer in both directions.
 * 
 * Wrap this in one of the specialised gRPC streaming patterns if you wish to simplify the 
 * API and not have to worry about when to call CloseSend, etc.
 */
export class HTTPgRPCWebSocket<ReqT extends ProtoJSONCompatible, ResT = any> {
	private parser: Parser<ResT>;
	private initialised: boolean = false;
	private open: Promise<void> = Promise.reject("socket is not yet open");
	private conn?: WebSocket = undefined;
	private responseBuffer: ResT[] = [];
	private mutex: IMutex = new Mutex();
	public readonly logger: Logger = console;
	public readonly name: string = "";
	public readonly id: string = uuid.v4();
	public readonly url: string;
	constructor(url: string, parser: Parser<ResT>, name?: string, logger?: Logger) {
		this.parser = parser;
		this.url = url;
		// Check undefined and wrong typing at the same time
		if (typeof name === "string") {
			this.name = name;
		}
		if (logger !== undefined) {
			this.logger = logger;
		}
	}
	public isInitialised(): Promise<boolean> {
		return this.mutex.Run(() => {
			return this.initialised;
		});
	}

	/** Wait until a message is successfully sent to the server. */
	public async Send(request: ReqT): Promise<void> {
		await this.open;
		return await this.mutex.RunAsync(async () => {
			await this.send(request);
		})
	}
	private async send(request: ReqT): Promise<void> {
		let message = request.ToProtoJSON();
		let messageString = JSON.stringify(message);
		this.conn?.send(messageString);
	}
	/** Wait until a message is received from the server. */
	public async Recv(): Promise<ResT> {
		await this.open;
		return await this.mutex.RunAsync(this.recv);
	}
	private async recv(): Promise<ResT> {
		let next = this.responseBuffer.shift();
		if (next === undefined) {
			// FIXME this isn't how buffers work, redo this whole section
			throw new Error("unimplemented")
		}
		return this.parser(next);
	}
	/** Close the sending direction of communications, any Send calls after this will throw an Error without writing to the websocket. */
	public async CloseSend(): Promise<void> {
		await this.open;
		this.conn?.send(EOFMessage);
	}
	/** Close terminates the websocket connection early, resulting in errors at the server's end. Any other calls after this point will throw an Error. */
	public async Close(code?: number, reason?: string): Promise<void> {
		await this.open;
		this.conn?.close(code, reason);
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
		return await this.mutex.RunAsync(this.initConnection);
	}
	private async initConnection(): Promise<EstablishedWebsocket<ReqT, ResT>> {
		if (this.initialised) {
			throw new Error("cannot initialise HTTPgRPCWebSocket twice");
		}

		this.initialised = true;
		let newConn = new WebSocket(this.url);
		this.conn = newConn;
		// Websocket opened without error, let's set up event listeners
		this.open = new Promise(async (resolve, reject) => {
			newConn.addEventListener("open", async ev => {
				let event = ev;
				try {
					await this.mutex.Run(() => {
						this.openHandler(event);
					})
					resolve();
				} catch (err) {
					this.logError(err, "caught error on open handler: ");
					reject(`error opening websocket: ${JSON.stringify(parseError(err))}`);
				}
			})
		})
		this.conn.addEventListener("close", async ev => {
			let event = ev;
			try {
				await this.mutex.Run(() => {
					this.closeHandler(event);
				})
			} catch (err) {
				this.logError(err, "caught error on close handler: ");
			}
		})
		this.conn.addEventListener("error", async ev => {
			let event = ev;
			try {
				await this.mutex.Run(() => {
					this.errorHandler(event);
				})
			} catch (err) {
				this.logError(err, "caught error on error handler: ");
			}
		})
		this.conn.addEventListener("message", async ev => {
			let event = ev;
			try {
				await this.mutex.Run(() => {
					this.messageHandler(event);
				})
			} catch (err) {
				this.logError(err, "caught error on message handler: ");
			}
		})
		return this;
	}

	//#region event handlers
	private async closeHandler(ev: CloseEvent): Promise<void> {
		let eventID = uuid.v4();
		let nameTag = this.nameTag(eventID);
		let result = `{code=${ev.code},reason=${ev.reason}}`;
		if (!ev.wasClean) {
			this.logger.warn(`${nameTag}closed uncleanly with ${result}`);
		} else {
			this.logger.log(`${nameTag}closed cleanly with ${result}`);
		}
		this.open = Promise.reject("socket has closed");
	}
	private async openHandler(ev: Event): Promise<void> {
		let eventID = uuid.v4();
		let nameTag = this.nameTag(eventID);
		this.logger.log(`${nameTag}opened`);
	}
	private async errorHandler(ev: Event): Promise<void> {
		let eventID = uuid.v4();
		let nameTag = this.nameTag(eventID);
		this.logger.error(`${nameTag}error on socket: `);
	}
	private async messageHandler(ev: MessageEvent<any>): Promise<void> {
		let parsed = await this.parser(ev.data);
		this.responseBuffer.push(parsed);
	}
	//#endregion event handlers

	//#region utility functions
	private logError(err: any, msgPrefix: string = "caught error: ", stackPrefix: string = "stack trace: ") {
		let eventID = uuid.v4();
		let name = this.nameTag(eventID);
		let parsed = parseError(err);
		this.logger.error(`${name}${msgPrefix}${parsed.msg}`);
		if (typeof parsed.stack === "string") {
			this.logger.log(`${name}${stackPrefix}${parsed.stack}`);
		}
	}
	private nameTag(eventID: string): string {
		return `WSS{id=${this.id},eventid=${eventID},name=${this.name}}: `;
	}
	//#endregion utility functions
}

interface parsedError {
	msg: string;
	stack?: string;
}

function parseError(err: any): parsedError {
	let msg: string;
	let stack: string | undefined = undefined;
	if (err instanceof Error) {
		msg = err.message;
		stack = err.stack;
	} else if (typeof err === "string") {
		msg = err;
	} else {
		msg = Object.assign(new Object(), err as Object).toString();
	}
	return {
		msg: msg,
		stack: stack
	}
}