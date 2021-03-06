import { assert } from 'chai';
import * as sinon from 'sinon';
import { ProtoJSONCompatible } from 'src/common';
import { HTTPgRPCWebSocket, IWebSocket } from './websocket';

class FakeMessage implements ProtoJSONCompatible {
	id?: string;
	num?: number;
	worked?: boolean;
	ToProtoJSON(): Object {
		return {
			id: this.id,
			num: this.num?.toString(),
			worked: this.worked
		}
	}
}

class FakeResponse {
	resultData?: string;
}

async function ParseFakeData(res: any): Promise<FakeResponse> {
	let parsed = JSON.parse(res);
	let resT: FakeResponse = {
		resultData: parsed.resultData
	};
	return resT;
}

function fakeWsConstructor(): IWebSocket {
	throw new Error("Unimplemented")
}

let fileSandbox = sinon.createSandbox();
before(async () => {
	global.WebSocket = fakeWsConstructor as any;
	fileSandbox.stub(global, "WebSocket").callsFake(fakeWsConstructor)
})
after(async () => {
	fileSandbox.restore();
})

describe("Websocket", () => {
	describe("Basic operations", () => {
		let sandbox: sinon.SinonSandbox;
		beforeEach(async () => {
			sandbox = sinon.createSandbox();
		})
		afterEach(async () => {
			sandbox.restore();
		})
		it("Init succeeds", async () => {
			let fake = new FakeWebsocket();
			let evStub = sandbox.stub(fake, "addEventListener");
			let evsCalled = [false, false, false, false];
			evStub.withArgs("close", sinon.match(() => true)).callsFake(() => {
				evsCalled[0] = true;
			})
			evStub.withArgs("open", sinon.match(() => true)).callsFake(() => {
				evsCalled[1] = true;
			})
			evStub.withArgs("message", sinon.match(() => true)).callsFake(() => {
				evsCalled[2] = true;
			})
			evStub.withArgs("error", sinon.match(() => true)).callsFake(() => {
				evsCalled[3] = true;
			})
			let ws = new HTTPgRPCWebSocket<FakeMessage, FakeResponse>("not a real URL", ParseFakeData, "TestWebsocket", console, () => fake);
			await ws.init();
			assert.deepEqual(evsCalled, [true, true, true, true]);
		});
		it.skip("Send and Recv succeed", async () => {
			let fake = new FakeWebsocket();
			let evStub = sandbox.stub(fake, "addEventListener");
			evStub.withArgs("close", sinon.match(() => true)).callsFake(() => {
			})
			evStub.withArgs("open", sinon.match(() => true)).callsFake(() => {
			})
			evStub.withArgs("message", sinon.match(() => true)).callsFake(() => {
			})
			evStub.withArgs("error", sinon.match(() => true)).callsFake(() => {
			})
			let ws = new HTTPgRPCWebSocket<FakeMessage, FakeResponse>("not a real URL", ParseFakeData, "TestWebsocket", console, () => fake);
			await ws.init();
			await ws.Send(new FakeMessage());
			let res = await ws.Recv();
			assert.equal(res, new FakeResponse());
		});
	})


})

class FakeWebsocket implements IWebSocket {
	send(data: string | ArrayBuffer | SharedArrayBuffer | Blob | ArrayBufferView): void {
		throw new Error("no websocket")
	}
	close(code?: number | undefined, reason?: string | undefined): void {
		throw new Error("no websocket")
	}
	addEventListener<K extends keyof WebSocketEventMap>(type: K, listener: (this: WebSocket, ev: WebSocketEventMap[K]) => any, options?: boolean | AddEventListenerOptions): void
	addEventListener(type: string, listener: EventListenerOrEventListenerObject, options?: boolean | AddEventListenerOptions): void {
		throw new Error("no websocket")
	}
}