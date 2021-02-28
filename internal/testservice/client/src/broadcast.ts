export function broadcast() {

	console.log("Opening websocket")
	let ws = new WebSocket("ws://127.0.0.1:4848/Broadcast");
	ws.addEventListener("message", async e => {
		console.log(`Received message: ${e.data}`)
	});
	ws.addEventListener("close", e => {
		console.log(`Websocket closed`)
	});
	ws.addEventListener("error", e => {
		console.log(`Received error: ${e}`)
	});
	ws.addEventListener("open", async e => {
		console.log(`Websocket opened`)
		console.log("Sending message")
		ws.send(`{"id":"totally an id"}`)
	});
}

export default broadcast;