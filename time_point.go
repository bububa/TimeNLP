package timenlp

import (
	"time"
)

// TimePoint 时间表达式单元规范化对应的内部类,
// 对应时间表达式规范化的每个字段，
// 六个字段分别是：年-月-日-时-分-秒，
// 每个字段初始化为-1
type TimePoint [6]int

// NewTimePointFromTime 基于时间新建TimePoint
func NewTimePointFromTime(t time.Time) TimePoint {
	return TimePoint{
		t.Year(),
		int(t.Month()),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
	}
}

// ToTime 转换为time.Time
func (t TimePoint) ToTime(loc *time.Location) time.Time {
	return time.Date(t[0], time.Month(t[1]), t[2], t[3], t[4], t[5], 0, loc)
}

// DefaultTimePoint 默认时间表达式单元
var DefaultTimePoint TimePoint = [6]int{-1, -1, -1, -1, -1, -1}
