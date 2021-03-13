import { FeedData, FeedResponse, FeedType } from "./service";
import { ClientStream, HTTPgRPCWebSocket } from "@llkennedy/httpgrpc"
import { base64 } from "rfc4648";

export async function feed() {
	let ws = new HTTPgRPCWebSocket<FeedData, FeedResponse>("ws://127.0.0.1:4848/Feed", FeedResponse.Parse, "FeedSocket", console);
	await ws.init()
	let stream = new ClientStream(ws);
	let data1 = new FeedData();
	data1.data_type = 998;
	data1.id = "totally an id";
	data1.raw_data = base64.parse("VGVzdCBEYXRh");
	data1.type = FeedType.FEED_TYPE_BLUE;
	let data2 = new FeedData();
	data2.data_type = 123;
	data2.id = "totally an id 2";
	data2.raw_data = base64.parse("VGVzdCBEYXRh");
	data2.type = FeedType.FEED_TYPE_RED;
	let data3 = new FeedData();
	data3.data_type = 654;
	data3.id = "totally an id 3";
	data3.raw_data = base64.parse("VGVzdCBEYXRh");
	await stream.Send(data1);
	await stream.Send(data2);
	await stream.Send(data3);
	try {
		let response = await stream.CloseAndRecv();
		console.log(JSON.stringify(response))
	} catch (err) {
		if (err instanceof Error) {
			console.error(`Caught error: ${err.message}`);
			console.debug(`Error trace: ${err.stack}`);
		} else {
			throw new Error('caught error was not error? ' + err);
		}
	}
}

export default feed;