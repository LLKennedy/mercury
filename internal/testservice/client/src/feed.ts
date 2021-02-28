import { sleep } from "@llkennedy/sleep.js"
import { EOFMessage } from "./constants";

export function feed() {

	var connected = false;

	console.log("Opening websocket")
	let ws = new WebSocket("ws://127.0.0.1:4848/Feed");
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
		console.log(`Websocket opened`)
		connected = true;
		console.log("Sending message")
		ws.send(`{"id":"totally an id","dataType":"998","rawData":"VGVzdCBEYXRh"}`)
		await sleep(1000);
		if (connected) {
			console.log("Sending message")
			ws.send(`{"id":"totally an id 2","dataType":"123","rawData":"VGVzdCBEYXRh"}`)
			await sleep(1000);
			if (connected) {
				console.log("Sending message")
				ws.send(`{"id":"totally an id 3","dataType":"654","rawData":"VGVzdCBEYXRh"}`)
				await sleep(1000);
				if (connected) {
					console.log("Sending EOF")
					ws.send(EOFMessage);
				}
			}
		}
	});
}

export default feed;