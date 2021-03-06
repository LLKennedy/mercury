import { assert } from 'chai';
import { sleep } from "@llkennedy/sleep.js";
import { IMutex, Mutex, SafeAction, SafeActionAsync } from "./Mutex"
import random from "seedrandom";

const countKey = "count";
const iterations = 100;

describe("Mutex", () => {
	describe("Fake Mutex", () => {
		it("Cannot handle simultaneous access", async () => {
			let f = new FakeMutex();
			// Without a mutex, only one read-write is completed
			assert.equal((await MutateMap(f)).get(countKey), 1);
		})
	})

	describe("Real Mutex", () => {
		it("Can handle many simultaneous access attempts", async () => {
			let m = new Mutex();
			// With a mutex, all operations are recorded
			assert.equal((await MutateMap(m)).get(countKey), iterations);
		});
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
	m.set(countKey, 0);
	let r = random(countKey);
	// This function will read, wait, then write
	// If anything else reads before it's done, the write will be lost when writing happens again
	const addOne = async () => {
		let start = m.get(countKey) ?? 0;
		await sleep(r.double() * 10)
		m.set(countKey, start + 1);
	}
	let jobs: Promise<void>[] = [];
	for (let i = 0; i < iterations; i++) {
		jobs.push(mutex.RunAsync(addOne))
	}
	await Promise.all(jobs);
	return m;
}
