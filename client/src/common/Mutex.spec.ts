import { assert } from 'chai';
import { sleep } from "@llkennedy/sleep.js";
import { IMutex, Mutex } from "./Mutex"

describe("Mutex", () => {
	describe("Multi-threading", () => {
		it("Can handle many simultaneous access attempts", async () => {
			let m: IMutex;
			m = new Mutex();
			let protectedArray: number[] = [1, 2];
			// TODO: make sure these examples would be obviously wrong in any other possible order
			const addFour = async () => {
				await sleep(1000);
				protectedArray.push(4);
			}
			const removeFirst = async () => {
				await sleep(100);
				protectedArray.shift();
			}
			const changeSecond = async () => {
				protectedArray[1] = 7;
			}
			let job1 = m.RunAsync(addFour);
			let job2 = m.RunAsync(removeFirst);
			let job3 = m.RunAsync(changeSecond);
			// Order here shouldn't matter
			await job3;
			await job2;
			await job1;
			assert.equal(protectedArray, [2, 7]);
		});
	})
});