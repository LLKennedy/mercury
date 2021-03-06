export type SafeAction = () => void;
export type SafeActionAsync = () => Promise<void>;

export interface IMutex {
	Run(codeToRun: SafeAction): Promise<void>;
	RunAsync(codeToRun: SafeActionAsync): Promise<void>;
}

export class Mutex implements IMutex {
	private current: Promise<void> = Promise.resolve();
	public async Run(codeToRun: SafeAction): Promise<void> {
		const next = async () => {
			codeToRun();
			return;
		}
		await this.RunAsync(next);
	}
	public async RunAsync(codeToRun: SafeActionAsync): Promise<void> {
		this.current = this.current.finally(codeToRun)
		await this.current;
	}
}