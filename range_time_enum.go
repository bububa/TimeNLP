package timenlp

// 范围时间的默认时间点
type RangeTimeEnum int

const (
	DAY_BREAK     RangeTimeEnum = 3  // 黎明
	EARLY_MORNING RangeTimeEnum = 8  // 早
	MORNING       RangeTimeEnum = 10 // 上午
	NOON          RangeTimeEnum = 12 // 中午、午间
	AFTERNOON     RangeTimeEnum = 15 // 下午、午后
	NIGHT         RangeTimeEnum = 18 // 晚上、傍晚
	LATE_NIGHT    RangeTimeEnum = 20 // 晚、晚间
	MID_NIGHT     RangeTimeEnum = 23 // 深夜
)
