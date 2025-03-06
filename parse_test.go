package timenlp

import (
	"testing"
	"time"
)

var (
	timeBase = time.Now()
	loc      = timeBase.Location()
)

// TestTimepoint 测试时间点
func TestTimepoint(t *testing.T) {
	normalizer := NewTimeNormalizer(true)
	target := "晚上8点到上午10点之间"
	t.Log(target)
	expectType := SPAN
	expectPoints := []time.Time{
		time.Date(timeBase.Year(), timeBase.Month(), timeBase.Day(), 20, 0, 0, 0, loc),
		time.Date(timeBase.AddDate(0, 0, 1).Year(), timeBase.AddDate(0, 0, 1).Month(), timeBase.AddDate(0, 0, 1).Day(), 10, 0, 0, 0, loc),
	}
	ret, err := normalizer.Parse(target, timeBase)
	if ret != nil {
		t.Log(ret.NormalizedString)
	}
	if err != nil {
		t.Error(err)
	} else if ret.Type != expectType {
		t.Errorf("expect: %s, got: %s", expectType, ret.Type)
	} else if len(ret.Points) != len(expectPoints) {
		t.Errorf("expect: %d points, result: %d points", len(expectPoints), len(ret.Points))
	} else {
		for idx, v := range ret.Points {
			if !v.Time.Equal(expectPoints[idx]) {
				t.Errorf("expect: %v, got: %v", expectPoints[idx], v)
			}
		}
	}
}

// TestTimeInChinese 测试中文时间
func TestTimeInChinese(t *testing.T) {
	normalizer := NewTimeNormalizer(true)
	target := "2013年二月二十八日下午四点三十分二十九秒"
	t.Log(target)
	expectType := TIMESTAMP
	expectPoints := []time.Time{
		time.Date(2013, 2, 28, 16, 30, 29, 0, loc),
	}
	ret, err := normalizer.Parse(target, timeBase)
	if ret != nil {
		t.Log(ret.NormalizedString)
	}
	if err != nil {
		t.Error(err)
	} else if ret.Type != expectType {
		t.Errorf("expect: %s, got: %s", expectType, ret.Type)
	} else if len(ret.Points) != len(expectPoints) {
		t.Errorf("expect: %d points, result: %d points", len(expectPoints), len(ret.Points))
	} else {
		for idx, v := range ret.Points {
			if !v.Time.Equal(expectPoints[idx]) {
				t.Errorf("expect: %v, got: %v", expectPoints[idx], v)
			}
		}
	}
}

// TestTimedelta 测试时间span
func TestTimedelta(t *testing.T) {
	normalizer := NewTimeNormalizer(true)
	ts := timeBase.AddDate(0, 0, 33).Add(2*time.Minute + 4*time.Second)
	target := "我需要大概33天2分钟四秒"
	t.Log(target)
	expectType := DELTA
	expectPoints := []time.Time{
		time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second(), 0, loc),
	}
	ret, err := normalizer.Parse(target, timeBase)
	if ret != nil {
		t.Log(ret.NormalizedString)
	}
	if err != nil {
		t.Error(err)
		return
	} else if ret.Type != expectType {
		t.Errorf("expect: %s, got: %s", expectType, ret.Type)
	} else if len(ret.Points) != len(expectPoints) {
		t.Errorf("expect: %d points, result: %d points", len(expectPoints), len(ret.Points))
	} else {
		for idx, v := range ret.Points {
			if !v.Time.Equal(expectPoints[idx]) {
				t.Errorf("expect: %v, got: %v", expectPoints[idx], v)
			}
		}
	}
}

// TestHoliday 测试节假日
func TestHoliday(t *testing.T) {
	normalizer := NewTimeNormalizer(true)
	target := "今年儿童节晚上九点一刻"
	t.Log(target)
	expectType := TIMESTAMP
	expectPoints := []time.Time{
		time.Date(timeBase.Year(), 6, 1, 21, 15, 0, 0, loc),
	}
	ret, err := normalizer.Parse(target, timeBase)
	if ret != nil {
		t.Log(ret.NormalizedString)
	}
	if err != nil {
		t.Error(err)
	} else if ret.Type != expectType {
		t.Errorf("expect: %s, got: %s", expectType, ret.Type)
	} else if len(ret.Points) != len(expectPoints) {
		t.Errorf("expect: %d points, result: %d points", len(expectPoints), len(ret.Points))
	} else {
		for idx, v := range ret.Points {
			if !v.Time.Equal(expectPoints[idx]) {
				t.Errorf("expect: %v, got: %v", expectPoints[idx], v)
			}
		}
	}
}

// TestDate 测试日期
func TestDate(t *testing.T) {
	normalizer := NewTimeNormalizer(false)
	target := "三日"
	t.Log(target)
	expectType := TIMESTAMP
	expectPoints := []time.Time{
		time.Date(timeBase.Year(), timeBase.Month(), 3, 0, 0, 0, 0, loc),
	}
	ret, err := normalizer.Parse(target, timeBase)
	if ret != nil {
		t.Log(ret.NormalizedString)
	}
	if err != nil {
		t.Error(err)
	} else if ret.Type != expectType {
		t.Errorf("expect: %s, got: %s", expectType, ret.Type)
	} else if len(ret.Points) != len(expectPoints) {
		t.Errorf("expect: %d points, result: %d points", len(expectPoints), len(ret.Points))
	} else {
		for idx, v := range ret.Points {
			if !v.Time.Equal(expectPoints[idx]) {
				t.Errorf("expect: %v, got: %v", expectPoints[idx], v)
			}
		}
	}
}

// TestLooseTime 测试宽松格式时间
func TestLooseTime(t *testing.T) {
	normalizer := NewTimeNormalizer(false)
	target := "7点4"
	t.Log(target)
	expectType := TIMESTAMP
	expectPoints := []time.Time{
		time.Date(timeBase.Year(), timeBase.Month(), timeBase.Day(), 7, 4, 0, 0, loc),
	}
	ret, err := normalizer.Parse(target, timeBase)
	if ret != nil {
		t.Log(ret.NormalizedString)
	}
	if err != nil {
		t.Error(err)
	} else if ret.Type != expectType {
		t.Errorf("expect: %s, got: %s", expectType, ret.Type)
	} else if len(ret.Points) != len(expectPoints) {
		t.Errorf("expect: %d points, result: %d points", len(expectPoints), len(ret.Points))
	} else {
		for idx, v := range ret.Points {
			if !v.Time.Equal(expectPoints[idx]) {
				t.Errorf("expect: %v, got: %v", expectPoints[idx], v)
			}
		}
	}
}

// TestSeason 测试节气
func TestSeason(t *testing.T) {
	normalizer := NewTimeNormalizer(true)
	targets := []string{"今年春分", "大年初一", "明年初一"}
	expectType := TIMESTAMP
	expectPoints := []time.Time{
		time.Date(timeBase.Year(), 3, 20, 0, 0, 0, 0, loc),
		time.Date(timeBase.Year(), 1, 29, 0, 0, 0, 0, loc),
		time.Date(timeBase.Year()+1, 2, 17, 0, 0, 0, 0, loc),
	}
	for idx, target := range targets {
		t.Log(target)
		ret, err := normalizer.Parse(target, timeBase)
		if ret != nil {
			t.Log(ret.NormalizedString)
		}
		if err != nil {
			t.Error(err)
		} else if ret.Type != expectType {
			t.Errorf("expect: %s, got: %s", expectType, ret.Type)
		} else if len(ret.Points) != 1 {
			t.Errorf("expect: 1 points, result: %d points", len(ret.Points))
		} else if !ret.Points[0].Time.Equal(expectPoints[idx]) {
			t.Errorf("expect: %v, got: %v", expectPoints[idx], ret.Points[0])
		}
	}
}

// TestPass10m 测试未来时间
func TestPass10m(t *testing.T) {
	normalizer := NewTimeNormalizer(true)
	ts := timeBase.Add(10 * time.Minute)
	target := "过十分钟"
	t.Log(target)
	expectType := DELTA
	expectPoints := []time.Time{
		time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second(), 0, loc),
	}
	ret, err := normalizer.Parse(target, timeBase)
	if ret != nil {
		t.Log(ret.NormalizedString)
	}
	if err != nil {
		t.Error(err)
	} else if ret.Type != expectType {
		t.Errorf("expect: %s, got: %s", expectType, ret.Type)
	} else if len(ret.Points) != len(expectPoints) {
		t.Errorf("expect: %d points, result: %d points", len(expectPoints), len(ret.Points))
	} else {
		for idx, v := range ret.Points {
			if !v.Time.Equal(expectPoints[idx]) {
				t.Errorf("expect: %v, got: %v", expectPoints[idx], v)
			}
		}
	}
}

// Test2hBefore 测试过去时间
func Test2hBefore(t *testing.T) {
	normalizer := NewTimeNormalizer(true)
	ts := timeBase.Add(-2 * time.Hour)
	target := "2个小时以前"
	t.Log(target)
	expectType := TIMESTAMP
	expectPoints := []time.Time{
		time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour(), 0, 0, 0, loc),
	}
	ret, err := normalizer.Parse(target, timeBase)
	if ret != nil {
		t.Log(ret.NormalizedString)
	}
	if err != nil {
		t.Error(err)
	} else if ret.Type != expectType {
		t.Errorf("expect: %s, got: %s", expectType, ret.Type)
	} else if len(ret.Points) != len(expectPoints) {
		t.Errorf("expect: %d points, result: %d points", len(expectPoints), len(ret.Points))
	} else {
		for idx, v := range ret.Points {
			if !v.Time.Equal(expectPoints[idx]) {
				t.Errorf("expect: %v, got: %v", expectPoints[idx], v)
			}
		}
	}
}

// TestNextMonday15m 测试未来时间
func TestNextMonday15m(t *testing.T) {
	normalizer := NewTimeNormalizer(true)
	ts := timeBase.AddDate(0, 0, int(7+1-timeBase.Weekday()))
	target := "Hi，all.下周一下午三点开会"
	t.Log(target)
	expectType := TIMESTAMP
	expectPoints := []time.Time{
		time.Date(ts.Year(), ts.Month(), ts.Day(), 15, 0, 0, 0, loc),
	}
	ret, err := normalizer.Parse(target, timeBase)
	if ret != nil {
		t.Log(ret.NormalizedString)
	}
	if err != nil {
		t.Error(err)
	} else if ret.Type != expectType {
		t.Errorf("expect: %s, got: %s", expectType, ret.Type)
	} else if len(ret.Points) != len(expectPoints) {
		t.Errorf("expect: %d points, result: %d points", len(expectPoints), len(ret.Points))
	} else {
		for idx, v := range ret.Points {
			if !v.Time.Equal(expectPoints[idx]) {
				t.Errorf("expect: %v, got: %v", expectPoints[idx], v)
			}
		}
	}
}

// TestMorning6m 测试早晚时间
func TestMorning6m(t *testing.T) {
	normalizer := NewTimeNormalizer(false)
	target := "早上六点起床"
	t.Log(target)
	expectType := TIMESTAMP
	expectPoints := []time.Time{
		time.Date(timeBase.Year(), timeBase.Month(), timeBase.Day(), 6, 0, 0, 0, loc),
	}
	ret, err := normalizer.Parse(target, timeBase)
	if ret != nil {
		t.Log(ret.NormalizedString)
	}
	if err != nil {
		t.Error(err)
	} else if ret.Type != expectType {
		t.Errorf("expect: %s, got: %s", expectType, ret.Type)
	} else if len(ret.Points) != len(expectPoints) {
		t.Errorf("expect: %d points, result: %d points", len(expectPoints), len(ret.Points))
	} else {
		for idx, v := range ret.Points {
			if !v.Time.Equal(expectPoints[idx]) {
				t.Errorf("expect: %v, got: %v", expectPoints[idx], v)
			}
		}
	}
}

// TestWeekday 测试星期几
func TestWeekday(t *testing.T) {
	normalizer := NewTimeNormalizer(true)
	ts := timeBase.AddDate(0, 0, int(7+1-timeBase.Weekday()))
	if timeBase.Weekday() == time.Monday {
		ts = timeBase
	}
	target := "周一开会"
	t.Log(target)
	expectType := TIMESTAMP
	expectPoints := []time.Time{
		time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, loc),
	}
	ret, err := normalizer.Parse(target, timeBase)
	if ret != nil {
		t.Log(ret.NormalizedString)
	}
	if err != nil {
		t.Error(err)
	} else if ret.Type != expectType {
		t.Errorf("expect: %s, got: %s", expectType, ret.Type)
	} else if len(ret.Points) != len(expectPoints) {
		t.Errorf("expect: %d points, result: %d points", len(expectPoints), len(ret.Points))
	} else {
		for idx, v := range ret.Points {
			if !v.Time.Equal(expectPoints[idx]) {
				t.Errorf("expect: %v, got: %v", expectPoints[idx], v)
			}
		}
	}
}

// TestNextNextMonday 测试未来星期几
func TestNextNextMonday(t *testing.T) {
	normalizer := NewTimeNormalizer(true)
	ts := timeBase.AddDate(0, 0, int(14+1-timeBase.Weekday()))
	target := "下下周一开会"
	t.Log(target)
	expectType := TIMESTAMP
	expectPoints := []time.Time{
		time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, loc),
	}
	ret, err := normalizer.Parse(target, timeBase)
	if ret != nil {
		t.Log(ret.NormalizedString)
	}
	if err != nil {
		t.Error(err)
	} else if ret.Type != expectType {
		t.Errorf("expect: %s, got: %s", expectType, ret.Type)
	} else if len(ret.Points) != len(expectPoints) {
		t.Errorf("expect: %d points, result: %d points", len(expectPoints), len(ret.Points))
	} else {
		for idx, v := range ret.Points {
			if !v.Time.Equal(expectPoints[idx]) {
				t.Errorf("expect: %v, got: %v", expectPoints[idx], v)
			}
		}
	}
}

// TestThisMaondayToNextMonday 测试星期几时间段
func TestThisMondayToNextMonday(t *testing.T) {
	normalizer := NewTimeNormalizer(false)
	ts := timeBase.AddDate(0, 0, -1*int(timeBase.Weekday()))
	ts2 := timeBase.AddDate(0, 0, int(7-timeBase.Weekday()))
	target := "本周日到下周日出差"
	t.Log(target)
	expectType := SPAN
	expectPoints := []time.Time{
		time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, loc),
		time.Date(ts2.Year(), ts2.Month(), ts2.Day(), 0, 0, 0, 0, loc),
	}
	ret, err := normalizer.Parse(target, timeBase)
	if ret != nil {
		t.Log(ret.NormalizedString)
	}
	if err != nil {
		t.Error(err)
	} else if ret.Type != expectType {
		t.Errorf("expect: %s, got: %s", expectType, ret.Type)
	} else if len(ret.Points) != len(expectPoints) {
		t.Errorf("expect: %d points, result: %d points", len(expectPoints), len(ret.Points))
	} else {
		for idx, v := range ret.Points {
			if !v.Time.Equal(expectPoints[idx]) {
				t.Errorf("expect: %v, got: %v", expectPoints[idx], v)
			}
		}
	}
}

// TestTimeSpanContext 测试时间段上下文
func TestTimeSpanContext(t *testing.T) {
	normalizer := NewTimeNormalizer(true)
	ts := timeBase.AddDate(0, 0, int(7+4-timeBase.Weekday()))
	if timeBase.Weekday() <= time.Thursday {
		ts = timeBase.AddDate(0, 0, int(4-timeBase.Weekday()))
	}
	target := "周四下午三点到五点开会"
	t.Log(target)
	expectType := SPAN
	expectPoints := []time.Time{
		time.Date(ts.Year(), ts.Month(), ts.Day(), 15, 0, 0, 0, loc),
		time.Date(ts.Year(), ts.Month(), ts.Day(), 17, 0, 0, 0, loc),
	}
	ret, err := normalizer.Parse(target, timeBase)
	if ret != nil {
		t.Log(ret.NormalizedString)
	}
	if err != nil {
		t.Error(err)
	} else if ret.Type != expectType {
		t.Errorf("expect: %s, got: %s", expectType, ret.Type)
	} else if len(ret.Points) != len(expectPoints) {
		t.Errorf("expect: %d points, result: %d points", len(expectPoints), len(ret.Points))
	} else {
		for idx, v := range ret.Points {
			if !v.Time.Equal(expectPoints[idx]) {
				t.Errorf("expect: %v, got: %v", expectPoints[idx], v)
			}
		}
	}
}

// TestStrictTime 测试严谨格式时间
func TestStrictTime(t *testing.T) {
	normalizer := NewTimeNormalizer(false)
	target := "6:30 起床"
	t.Log(target)
	expectType := TIMESTAMP
	expectPoints := []time.Time{
		time.Date(timeBase.Year(), timeBase.Month(), timeBase.Day(), 6, 30, 0, 0, loc),
	}
	ret, err := normalizer.Parse(target, timeBase)
	if ret != nil {
		t.Log(ret.NormalizedString)
	}
	if err != nil {
		t.Error(err)
	} else if ret.Type != expectType {
		t.Errorf("expect: %s, got: %s", expectType, ret.Type)
	} else if len(ret.Points) != len(expectPoints) {
		t.Errorf("expect: %d points, result: %d points", len(expectPoints), len(ret.Points))
	} else {
		for idx, v := range ret.Points {
			if !v.Time.Equal(expectPoints[idx]) {
				t.Errorf("expect: %v, got: %v", expectPoints[idx], v)
			}
		}
	}
}

// TestTomorrowMorning 测试明天早晚
func TestTomorrowMorning(t *testing.T) {
	normalizer := NewTimeNormalizer(true)
	ts := timeBase.AddDate(0, 0, 1)
	target := "明天早上跑步"
	t.Log(target)
	expectType := TIMESTAMP
	expectPoints := []time.Time{
		time.Date(ts.Year(), ts.Month(), ts.Day(), int(EARLY_MORNING), 0, 0, 0, loc),
	}
	ret, err := normalizer.Parse(target, timeBase)
	if ret != nil {
		t.Log(ret.NormalizedString)
	}
	if err != nil {
		t.Error(err)
	} else if ret.Type != expectType {
		t.Errorf("expect: %s, got: %s", expectType, ret.Type)
	} else if len(ret.Points) != len(expectPoints) {
		t.Errorf("expect: %d points, result: %d points", len(expectPoints), len(ret.Points))
	} else {
		for idx, v := range ret.Points {
			if !v.Time.Equal(expectPoints[idx]) {
				t.Errorf("expect: %v, got: %v", expectPoints[idx], v)
			}
		}
	}
}

// TestLongText1 测试长文字
func TestLongText1(t *testing.T) {
	normalizer := NewTimeNormalizer(false)
	target := `7月 10日晚上7 点左右，六安市公安局裕安分局平桥派出所接到辖区居民戴某报警称，到同学家玩耍的女儿迟迟未归，手机也打不通了。很快，派出所又接到与戴某同住一小区的王女士报警：下午5点左右，12岁的儿子和同学在家中吃过晚饭后，带着3 岁的弟弟一起出了门，之后便没了消息，手机也关机了。短时间内，接到两起孩子失联的报警，值班民警张晖和队友立即前往小区。`
	t.Log(target)
	expectType := SPAN
	expectPoints := []time.Time{
		time.Date(timeBase.Year(), 7, 10, 19, 0, 0, 0, loc),
		time.Date(timeBase.Year(), 7, 10, 17, 0, 0, 0, loc),
	}
	ret, err := normalizer.Parse(target, timeBase)
	if ret != nil {
		t.Log(ret.NormalizedString)
	}
	if err != nil {
		t.Error(err)
	} else if ret.Type != expectType {
		t.Errorf("expect: %s, got: %s", expectType, ret.Type)
	} else if len(ret.Points) != len(expectPoints) {
		t.Errorf("expect: %d points, result: %d points", len(expectPoints), len(ret.Points))
		for _, v := range ret.Points {
			t.Logf("got: %v", v)
		}
	} else {
		for idx, v := range ret.Points {
			if !v.Time.Equal(expectPoints[idx]) {
				t.Errorf("expect: %v, got: %v", expectPoints[idx], v)
			}
		}
	}
}

// TestLongText2 测试长文字
func TestLongText2(t *testing.T) {
	normalizer := NewTimeNormalizer(false)
	target := `《辽宁日报》今日报道，7月18日辽宁召开省委常委扩大会，会议从下午两点半开到六点半，主要议题为：落实中央巡视整改要求。`
	t.Log(target)
	expectType := SPAN
	expectPoints := []time.Time{
		time.Date(timeBase.Year(), timeBase.Month(), timeBase.Day(), 0, 0, 0, 0, loc),
		time.Date(timeBase.Year(), 7, 18, 0, 0, 0, 0, loc),
		time.Date(timeBase.Year(), 7, 18, 14, 30, 0, 0, loc),
		time.Date(timeBase.Year(), 7, 18, 18, 30, 0, 0, loc),
	}
	ret, err := normalizer.Parse(target, timeBase)
	if ret != nil {
		t.Log(ret.NormalizedString)
	}
	if err != nil {
		t.Error(err)
	} else if ret.Type != expectType {
		t.Errorf("expect: %s, got: %s", expectType, ret.Type)
	} else if len(ret.Points) != len(expectPoints) {
		t.Errorf("expect: %d points, result: %d points", len(expectPoints), len(ret.Points))
		for _, v := range ret.Points {
			t.Logf("got: %v", v)
		}
	} else {
		for idx, v := range ret.Points {
			if !v.Time.Equal(expectPoints[idx]) {
				t.Errorf("expect: %v, got: %v", expectPoints[idx], v)
			}
		}
	}
}

// TestLooseDate 测试宽松格式日期
// * test failed
func TestLooseDate(t *testing.T) {
	t.Skip("Skipping testing in CI environment")
	normalizer := NewTimeNormalizer(false)
	target := "6-3 春游"
	t.Log(target)
	expectType := TIMESTAMP
	expectPoints := []time.Time{
		time.Date(timeBase.Year(), 6, 3, 0, 0, 0, 0, loc),
	}
	ret, err := normalizer.Parse(target, timeBase)
	if ret != nil {
		t.Log(ret.NormalizedString)
	}
	if err != nil {
		t.Error(err)
	} else if ret.Type != expectType {
		t.Errorf("expect: %s, got: %s", expectType, ret.Type)
	} else if len(ret.Points) != len(expectPoints) {
		t.Errorf("expect: %d points, result: %d points", len(expectPoints), len(ret.Points))
	} else {
		for idx, v := range ret.Points {
			if !v.Time.Equal(expectPoints[idx]) {
				t.Errorf("expect: %v, got: %v", expectPoints[idx], v)
			}
		}
	}
}

// TestWeeks 解析某年某月的第n周
func TestWeeks(t *testing.T) {
	normalizer := NewTimeNormalizer(false)
	target := "Hi，all.2025年11月第三周"
	t.Log(target)
	expectType := TIMESTAMP
	expectPoints := []time.Time{
		time.Date(2025, 11, 15, 0, 0, 0, 0, loc),
	}
	ret, err := normalizer.Parse(target, timeBase)
	if ret != nil {
		t.Log(ret.NormalizedString)
	}
	if err != nil {
		t.Error(err)
	} else if ret.Type != expectType {
		t.Errorf("expect: %s, got: %s", expectType, ret.Type)
	} else if len(ret.Points) != len(expectPoints) {
		t.Errorf("expect: %d points, result: %d points", len(expectPoints), len(ret.Points))
	} else {
		for idx, v := range ret.Points {
			if !v.Time.Equal(expectPoints[idx]) {
				t.Errorf("expect: %v, got: %v", expectPoints[idx], v)
			}
		}
	}
}

// TestSeasons 解析某年第n季
func TestSeasons(t *testing.T) {
	normalizer := NewTimeNormalizer(true)
	target := "Hi，all.2025年第二季"
	t.Log(target)
	expectType := TIMESTAMP
	expectPoints := []time.Time{
		time.Date(2025, 4, 1, 0, 0, 0, 0, loc),
	}
	ret, err := normalizer.Parse(target, timeBase)
	if ret != nil {
		t.Log(ret.NormalizedString)
	}
	if err != nil {
		t.Error(err)
	} else if ret.Type != expectType {
		t.Errorf("expect: %s, got: %s", expectType, ret.Type)
	} else if len(ret.Points) != len(expectPoints) {
		t.Errorf("expect: %d points, result: %d points", len(expectPoints), len(ret.Points))
	} else {
		for idx, v := range ret.Points {
			if !v.Time.Equal(expectPoints[idx]) {
				t.Errorf("expect: %v, got: %v", expectPoints[idx], v)
			}
		}
	}
}
