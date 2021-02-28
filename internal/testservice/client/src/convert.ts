import { sleep } from "@llkennedy/sleep.js";

export function convert() {
	var connected = false;

	console.log("Opening websocket")
	let ws = new WebSocket("ws://127.0.0.1:4848/ConvertToString");
	ws.addEventListener("message", async e => {
		console.log(`Received message: ${e.data}`)
	});
	ws.addEventListener("close", e => {
		console.log(`Websocket closed`)
		connected = false;
	});
	ws.addEventListener("error", e => {
		console.log(`Received error: ${e}`)
	});
	ws.addEventListener("open", async e => {
		connected = true;
		console.log(`Websocket opened`)
		console.log("Sending message")
		while (connected) {
			await sleep(1000);
			ws.send(`{"rawData":"MQ=="}`)
		}
	});
}

export default convert;