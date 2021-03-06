export type SafeAction = () => void;
export type SafeActionAsync = () => Promise<void>;

export interface IMutex {
	Run(codeToRun: SafeAction): Promise<void>;
	RunAsync(codeToRun: SafeActionAsync): Promise<void>;
}

export class Mutex implements IMutex {
	private current: Promise<void> = Promise.resolve();
	public async Run(codeToRun: SafeAction): Promise<void> {
		await this.current;
		// FIXME: mark new current
		codeToRun();
		// FIXME: resolve new current
	}
	public async RunAsync(codeToRun: SafeActionAsync): Promise<void> {
		await this.current;
		// FIXME: mark new current
		await codeToRun();
		// FIXME: resolve new current
	}
}