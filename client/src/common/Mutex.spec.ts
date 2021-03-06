import { assert } from 'chai';
import { sleep } from "@llkennedy/sleep.js";
import { IMutex, Mutex, SafeAction, SafeActionAsync } from "./Mutex"
import random from "seedrandom";

describe("Mutex", () => {
	describe("Fake Mutex", () => {
		it("Cannot handle simultaneous access", async () => {
			let f = new FakeMutex();
			assert.deepEqual(await MutateMap(f), new Map<string, number>());
		})
	})

	describe("Real Mutex", () => {
		describe("Multi-threading", () => {
			it("Can handle many simultaneous access attempts", async () => {
				let m = new Mutex();
				assert.deepEqual(await MutateMap(m), new Map<string, number>());
			});
		})
	});
})

class FakeMutex implements IMutex {
	async Run(codeToRun: SafeAction): Promise<void> {
		return codeToRun();
	}
	async RunAsync(codeToRun: SafeActionAsync): Promise<void> {
		return codeToRun();
	}
}

async function MutateMap(mutex: IMutex): Promise<Map<string, number>> {
	let m = new Map<string, number>();
	let r = random("test")
	await sleep(r.double() * 10);
	return m;
}