interface enumMap1 {
	[inNum: number]: string;
};
interface enumMap2 {
	[inStr: string]: number;
}
export type EnumMap = enumMap1 & enumMap2;