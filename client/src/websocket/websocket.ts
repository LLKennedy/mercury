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

/** EOFError indicates that the server closed the channel cleanly after transmitting all it intended to */
export class EOFError extends Error {
	constructor() {
		super(EOFMessage);
	}
}

/** A websocket connection open and ready to handle gRPC message transfer in both directions.
 * 
 * Wrap this in one of the specialised gRPC streaming patterns if you wish to simplify the 
 * API and not have to worry about when to call CloseSend, etc.
 */
export class HTTPgRPCWebSocket<ReqT extends ProtoJSONCompatible, ResT = any> {
	/** The logger used for all logs from this class */
	public readonly logger: Logger = console;
	/** The name of this websocket */
	public readonly name: string = "";
	/** The UUID of this websocket */
	public readonly id: string = uuid.v4();
	/** The URL for this websocket */
	public readonly url: string;
	/** The function this websocket uses to parse response data into the desired response message class */
	public readonly parser: Parser<ResT>;
	/** Whether or not this class has been properly set up by its init() function */
	private initialised: boolean = false;
	/** Rejects if sending is not yet ready or has been closed after opening */
	private sendOpen: Promise<void> = Promise.reject("socket is not yet open");
	/** Rejects if receiving is not yet ready or has been closed after opening */
	private recvOpen: Promise<void> = Promise.reject("socket is not yet open");
	/** The raw connection object, or a placeholder if the connection is not yet open */
	private conn: IWebsocket = new NoWebsocket();
	/** Parsed responses waiting to be read by Recv calls */
	private responseBuffer: ResT[] = [];
	/** Core mutex for init calls */
	private mutex: IMutex = new Mutex();
	/** Mutex for send operations, independent from recvMutex */
	private sendMutex: IMutex = new Mutex();
	/** Mutex for receive operations, independent from sendMutex */
	private recvMutex: IMutex = new Mutex();
	/** Resolves when a message arrives */
	private messageAlert: Promise<void> = Promise.reject("socket is not yet open");
	/** For use *only* by the message handler */
	private messageArrived: () => void = () => { };
	private messageFailed: (err: any) => void = () => { };

	/** Constructor */
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

	/** Wait until a message is successfully sent to the server. */
	public async Send(request: ReqT): Promise<void> {
		await this.sendOpen;
		return await this.sendMutex.RunAsync(async () => {
			await this.send(request);
		})
	}
	private async send(request: ReqT): Promise<void> {
		let message = request.ToProtoJSON();
		let messageString = JSON.stringify(message);
		this.conn.send(messageString);
	}

	/** Wait until a message is received from the server. */
	public async Recv(): Promise<ResT> {
		let finishing = false;
		try {
			await this.recvOpen;
		} catch (err) {
			// EOF error indicates the socket is closed, but there may be already-parsed messages still to receive
			if (!(err instanceof EOFError)) {
				// Any other error, just re-throw, you don't get to parse your data
				throw err;
			} else {
				finishing = true;
			}
		}
		return await this.recv(finishing);
	}
	private async recv(finishing: boolean): Promise<ResT> {
		let next: ResT | undefined;
		do { next = await this.recvGetOne() } while (next === undefined && !finishing) {
			try {
				await this.messageAlert;
			} catch (err) {
				if (err instanceof EOFError) {
					// closed but might still have a final message for us
					finishing = true;
				} else {
					throw err;
				}
			}
		}
		if (next === undefined) {
			// Only way to be here is if we got EOF and there are no more messages in the buffer
			throw new EOFError();
		}
		return next;
	}
	private async recvGetOne(): Promise<ResT | undefined> {
		return await this.recvMutex.Run(() => {
			return this.responseBuffer.shift();
		})
	}

	/** Close the sending direction of communications, any Send calls after this will throw an Error without writing to the websocket. */
	public async CloseSend(): Promise<void> {
		await this.sendOpen;
		await this.mutex.RunAsync(this.closeSend);
	}
	private async closeSend(): Promise<void> {
		this.conn.send(EOFMessage);
		this.sendOpen = Promise.reject("socket closed for sending");
	}

	/** Close terminates the websocket connection early, resulting in errors at the server's end. Any other calls after this point will throw an Error. */
	public async Close(code?: number, reason?: string): Promise<void> {
		await this.recvOpen;
		await this.sendMutex.RunAsync(async () => {
			await this.recvMutex.RunAsync(async () => {
				await this.close(code, reason);
			})
		});
	}
	private async close(code?: number, reason?: string): Promise<void> {
		this.conn.close(code, reason);
		this.sendOpen = Promise.reject("socket manually closed early");
		this.recvOpen = this.sendOpen;
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
		this.sendOpen = new Promise(async (resolve, reject) => {
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
		this.recvOpen = this.sendOpen;
		newConn.addEventListener("close", async ev => {
			let event = ev;
			try {
				await this.mutex.Run(() => {
					this.closeHandler(event);
				})
			} catch (err) {
				this.logError(err, "caught error on close handler: ");
			}
		})
		newConn.addEventListener("error", async ev => {
			let event = ev;
			try {
				await this.mutex.Run(() => {
					this.errorHandler(event);
				})
			} catch (err) {
				this.logError(err, "caught error on error handler: ");
			}
		})
		this.newMessageAlert();
		newConn.addEventListener("message", async ev => {
			let event = ev;
			try {
				await this.mutex.Run(() => {
					this.messageHandler(event);
					this.messageArrived();
				})
			} catch (err) {
				this.logError(err, "caught error on message handler: ");
				await this.mutex.Run(() => {
					this.messageFailed(err);
				})
			} finally {
				await this.mutex.Run(this.newMessageAlert)
			}
		})
		return this;
	}
	private newMessageAlert() {
		this.messageAlert = new Promise(async (resolve, reject) => {
			this.messageArrived = resolve;
			this.messageFailed = reject;
		});
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
		this.conn = new NoWebsocket();
		this.sendOpen = Promise.reject("socket has closed");
		this.recvOpen = ev.wasClean ? Promise.reject(new EOFError()) : this.sendOpen;
		this.messageFailed(new EOFError());
	}
	private async openHandler(ev: Event): Promise<void> {
		let eventID = uuid.v4();
		let nameTag = this.nameTag(eventID);
		this.logger.log(`${nameTag}opened`);
	}
	private async errorHandler(ev: Event): Promise<void> {
		this.logError(ev, "socket closed due to error: ");
		this.conn = new NoWebsocket();
		this.sendOpen = Promise.reject("socket has closed due to error");
		this.recvOpen = this.sendOpen;
		this.messageFailed(ev);
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

interface IWebsocket {
	send(data: string | ArrayBuffer | SharedArrayBuffer | Blob | ArrayBufferView): void;
	close(code?: number | undefined, reason?: string | undefined): void;
}

class NoWebsocket implements IWebsocket {
	send(data: string | ArrayBuffer | SharedArrayBuffer | Blob | ArrayBufferView): void {
		throw new Error("no websocket")
	}
	close(code?: number | undefined, reason?: string | undefined): void {
		throw new Error("no websocket")
	}
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