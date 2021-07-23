package timenlp

// RangeTimeEnum 范围时间的默认时间点
type RangeTimeEnum int

const (
	// DAY_BREAK 黎明
	DAY_BREAK RangeTimeEnum = 3
	// EARLY_MORNING 早
	EARLY_MORNING RangeTimeEnum = 8
	// MORNING 上午
	MORNING RangeTimeEnum = 10
	// NOON 中午、午间
	NOON RangeTimeEnum = 12
	// AFTERNOON 下午、午后
	AFTERNOON RangeTimeEnum = 15
	// NIGHT 晚上、傍晚
	NIGHT RangeTimeEnum = 18
	// LATE_NIGHT 晚、晚间
	LATE_NIGHT RangeTimeEnum = 20
	// MID_NIGHT 深夜
	MID_NIGHT RangeTimeEnum = 23
)
