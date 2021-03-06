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
		});
	})

	describe("Real Mutex", () => {
		it("Can handle many simultaneous access attempts", async () => {
			let m = new Mutex();
			// With a mutex, all operations are recorded
			assert.equal((await MutateMap(m)).get(countKey), iterations);
		});
		it("Can handle partial failures", async () => {
			let m = new Mutex();
			let job1 = m.RunAsync(async () => {
				sleep(100);
				return;
			})
			let job2 = m.RunAsync(async () => {
				sleep(10);
				throw new Error("failure");
			})
			let job3 = m.RunAsync(async () => {
				sleep(50);
				return;
			})
			let job4 = m.Run(() => {
				throw new Error("something's wrong");
			})
			try {
				await job1;
			} catch (err) {
				assert.fail(err)
			}
			try {
				await job2;
				assert.fail("Not supposed to succeed")
			} catch (err) {
				assert.instanceOf(err, Error)
				assert.equal((err as Error).message, "failure")
			}
			try {
				await job3;
			} catch (err) {
				assert.fail(err)
			}
			try {
				await job4;
				assert.fail("Not supposed to succeed")
			} catch (err) {
				assert.instanceOf(err, Error)
				assert.equal((err as Error).message, "something's wrong")
			}
		})
	});
})

class FakeMutex implements IMutex {
	async Run<T>(codeToRun: SafeAction<T>): Promise<T> {
		return codeToRun();
	}
	async RunAsync<T>(codeToRun: SafeActionAsync<T>): Promise<T> {
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
