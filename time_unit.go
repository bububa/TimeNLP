package timenlp

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dlclark/regexp2"
)

// TimeUnit 时间语句分析
type TimeUnit struct {
	expTime                 string
	normalizer              *TimeNormalizer
	tp                      TimePoint
	tpOrigin                TimePoint
	noYear                  bool
	isMorning               bool
	isAllDayTime            bool
	isFirstTimeSolveContext bool
	pos                     int
	length                  int
	ts                      time.Time
}

// NewTimeUnit 新建TimeUnit
func NewTimeUnit(expTime string, pos int, length int, normalizer *TimeNormalizer, tpCtx TimePoint) *TimeUnit {
	ret := &TimeUnit{
		expTime:                 expTime,
		normalizer:              normalizer,
		tp:                      DefaultTimePoint,
		tpOrigin:                tpCtx,
		isMorning:               false,
		isFirstTimeSolveContext: true,
		isAllDayTime:            true,
		pos:                     pos,
		length:                  length,
	}
	ret.normalization()
	return ret
}

// ToResultPoint 转换为ResultPoint
func (t TimeUnit) ToResultPoint() ResultPoint {
	return ResultPoint{
		Time:   t.Time(),
		Pos:    t.pos,
		Length: t.length,
	}
}

// Time 转换为time.Time
func (t TimeUnit) Time() time.Time {
	return t.ts
}

// normalization 标准化
func (t *TimeUnit) normalization() {
	t.normSetYear()
	t.normSetMonth()
	t.normSetDay()
	t.normSetMonthFuzzyDay()
	t.normSetBaseRelated()
	t.normSetCurRelated()
	t.normSetHour()
	t.normSetMinute()
	t.normSetSecond()
	t.normSetSpecial()
	t.normSetSpanRelated()
	t.normSetHoliday()
	t.normSetTotal()
	t.modifyTimeBase()
	for idx, v := range t.tp {
		t.tpOrigin[idx] = v
	}

	// 判断是时间点还是时间区间
	flag := true
	idx := 0
	for idx < 4 {
		if t.tp[idx] != -1 {
			flag = false
			break
		}
		idx += 1
	}
	if flag {
		t.normalizer.isTimeSpan = true
	}
	if t.normalizer.isTimeSpan {
		var days int64
		if t.tp[0] > 0 {
			days += int64(365 * t.tp[0])
		}
		if t.tp[1] > 0 {
			days += int64(30 * t.tp[1])
		}
		if t.tp[2] > 0 {
			days += int64(t.tp[2])
		}
		var tunit TimePoint
		idx = 3
		for idx < 6 {
			if t.tp[idx] < 0 {
				tunit[idx] = 0
			} else {
				tunit[idx] = t.tp[idx]
			}
			idx += 1
		}
		seconds := int64(tunit[3]*3600) + int64(tunit[4]*60) + int64(tunit[5])
		if seconds == 0 && days == 0 {
			t.normalizer.isTimeSpan = false
			t.normalizer.invalidSpan = true
			return
		}
		t.ts = t.normalizer.timeBase.Add(t.genSpan(days, seconds)).Truncate(time.Second)
		return
	}
	tunitPointer := 5
	for tunitPointer >= 0 && t.tp[tunitPointer] < 0 {
		tunitPointer -= 1
	}
	idx = 0
	timeGrid := NewTimePointFromTime(t.normalizer.timeBase)
	for idx < tunitPointer {
		if t.tp[idx] < 0 {
			t.tp[idx] = timeGrid[idx]
		}
		idx += 1
	}
	t.ts = t.genTime()
}

// genSpan 转化为time.Duration
func (t *TimeUnit) genSpan(days int64, second int64) time.Duration {
	return (time.Duration(days)*24*time.Hour + time.Duration(second)*time.Second)
}

// getTime 获取time.Time
func (t *TimeUnit) genTime() time.Time {
	var zero time.Time
	ret := NewTimePointFromTime(zero)
	for idx, v := range t.tp {
		if v > 0 {
			ret[idx] = v
		}
	}
	return ret.ToTime(t.normalizer.timeBase.Location())
}

// normSetYear 年-规范化方法--该方法识别时间表达式单元的年字段
func (t *TimeUnit) normSetYear() {
	// 一位数表示的年份
	{
		pattern := regexp2.MustCompile("(?<![0-9])[0-9]{1}(?=年)", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			t.normalizer.isTimeSpan = true
			year, _ := strconv.Atoi(match.String())
			t.tp[0] = year
		}
	}
	// 两位数表示的年份
	{
		pattern := regexp2.MustCompile("[0-9]{2}(?=年)", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			year, _ := strconv.Atoi(match.String())
			t.tp[0] = year
		}
	}
	// 三位数表示的年份
	{
		pattern := regexp2.MustCompile("(?<![0-9])[0-9]{3}(?=年)", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			t.normalizer.isTimeSpan = true
			year, _ := strconv.Atoi(match.String())
			t.tp[0] = year
		}
	}
	// 四位数表示的年份
	{
		pattern := regexp2.MustCompile("[0-9]{4}(?=年)", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			year, _ := strconv.Atoi(match.String())
			t.tp[0] = year
		}
	}
}

// normSetMonth 月-规范化方法--该方法识别时间表达式单元的月字段
func (t *TimeUnit) normSetMonth() {
	pattern := regexp2.MustCompile("((10)|(11)|(12)|([1-9]))(?=月)", 0)
	if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
		month, _ := strconv.Atoi(match.String())
		t.tp[1] = month
		t.preferFuture(1)
	}
}

// normSetDay 日-规范化方法：该方法识别时间表达式单元的日字段
func (t *TimeUnit) normSetDay() {
	pattern := regexp2.MustCompile("((?<!\\d))([0-3][0-9]|[1-9])(?=(日|号))", 0)
	if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
		day, _ := strconv.Atoi(match.String())
		t.tp[2] = day
		t.preferFuture(2)
		t.checkTime(t.tp)
	}
}

// normSetMonthFuzzyDay 月-日 兼容模糊写法：该方法识别时间表达式单元的月、日字段
func (t *TimeUnit) normSetMonthFuzzyDay() {
	pattern := regexp.MustCompile("((10)|(11)|(12)|([1-9]))(月|\\.|\\-)([0-3][0-9]|[1-9])")
	match := pattern.FindAllString(t.expTime, -1)
	for _, m := range match {
		p := regexp.MustCompile("(月|\\.|\\-)")
		if loc := p.FindStringIndex(m); loc != nil {
			month, _ := strconv.Atoi(m[0:loc[0]])
			day, _ := strconv.Atoi(m[loc[1]:])
			t.tp[1] = month
			t.tp[2] = day
			// 处理倾向于未来时间的情况
			t.preferFuture(1)
		}
		t.checkTime(t.tp)
	}
}

// normSetBaseRelated 设置以上文时间为基准的时间偏移计算
func (t *TimeUnit) normSetBaseRelated() {
	cur := time.Now()
	flag := []bool{false, false, false, false}
	{
		pattern := regexp2.MustCompile("\\d+(?=(个)?小时[以之]?前)", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			flag[3] = true
			hour, _ := strconv.Atoi(match.String())
			cur = cur.Add(time.Duration(-1*hour) * time.Hour)
		}
	}
	{
		pattern := regexp2.MustCompile("\\d+(?=(个)?小时[以之]?后)", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			flag[3] = true
			hour, _ := strconv.Atoi(match.String())
			cur = cur.Add(time.Duration(hour) * time.Hour)
		}
	}
	{
		pattern := regexp2.MustCompile("\\d+(?=天[以之]?前)", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			flag[2] = true
			days, _ := strconv.Atoi(match.String())
			cur = cur.AddDate(0, 0, -1*days)
		}
	}
	{
		pattern := regexp2.MustCompile("\\d+(?=天[以之]?后)", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			flag[2] = true
			days, _ := strconv.Atoi(match.String())
			cur = cur.AddDate(0, 0, days)
		}
	}
	{
		pattern := regexp2.MustCompile("\\d+(?=(个)?月[以之]?前)", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			flag[1] = true
			month, _ := strconv.Atoi(match.String())
			cur = cur.AddDate(0, -1*month, 0)
		}
	}
	{
		pattern := regexp2.MustCompile("\\d+(?=(个)?月[以之]?后)", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			flag[1] = true
			month, _ := strconv.Atoi(match.String())
			cur = cur.AddDate(0, month, 0)
		}
	}
	{
		pattern := regexp2.MustCompile("\\d+(?=年[以之]?前)", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			flag[0] = true
			years, _ := strconv.Atoi(match.String())
			cur = cur.AddDate(-1*years, 0, 0)
		}
	}
	{
		pattern := regexp2.MustCompile("\\d+(?=年[以之]?后)", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			flag[0] = true
			years, _ := strconv.Atoi(match.String())
			cur = cur.AddDate(years, 0, 0)
		}
	}
	if flag[0] || flag[1] || flag[2] || flag[3] {
		t.tp[0] = cur.Year()
	}
	if flag[1] || flag[2] || flag[3] {
		t.tp[1] = int(cur.Month())
	}
	if flag[2] || flag[3] {
		t.tp[2] = cur.Day()
	}
	if flag[3] {
		t.tp[3] = cur.Hour()
	}
}

// normSetCurRelated 设置当前时间相关的时间表达式
func (t *TimeUnit) normSetCurRelated() {
	cur := time.Now()
	flag := []bool{false, false, false}
	if strings.Contains(t.expTime, "前年") {
		flag[0] = true
		cur = cur.AddDate(-2, 0, 0)
	}
	if strings.Contains(t.expTime, "去年") {
		flag[0] = true
		cur = cur.AddDate(-1, 0, 0)
	}
	if strings.Contains(t.expTime, "今年") {
		flag[0] = true
	}
	if strings.Contains(t.expTime, "明年") {
		flag[0] = true
		cur = cur.AddDate(1, 0, 0)
	}
	if strings.Contains(t.expTime, "后年") {
		flag[0] = true
		cur = cur.AddDate(2, 0, 0)
	}
	{
		pattern := regexp.MustCompile(`上*上(个)?月`)
		if pattern.MatchString(t.expTime) {
			flag[1] = true
			cnt := strings.Count(t.expTime, "上")
			cur = cur.AddDate(0, -1*cnt, 0)
		}
	}
	{
		pattern := regexp.MustCompile(`(本|这个)月`)
		if pattern.MatchString(t.expTime) {
			flag[1] = true
		}
	}
	{
		pattern := regexp.MustCompile(`下*下(个)?月`)
		if pattern.MatchString(t.expTime) {
			flag[1] = true
			cnt := strings.Count(t.expTime, "下")
			cur = cur.AddDate(0, cnt, 0)
		}
	}
	{
		pattern := regexp.MustCompile(`大*大前天`)
		if pattern.MatchString(t.expTime) {
			flag[2] = true
			cnt := strings.Count(t.expTime, "大")
			cur = cur.AddDate(0, 0, -1*cnt)
		}
	}
	{
		pattern := regexp2.MustCompile("(?<!大)前天", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			flag[2] = true
			cur = cur.AddDate(0, 0, -2)
		}
	}
	if strings.Contains(t.expTime, "昨") {
		flag[2] = true
		cur = cur.AddDate(0, 0, -1)
	}
	{
		pattern := regexp2.MustCompile("今(?!年)", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			flag[2] = true
		}
	}
	{
		pattern := regexp2.MustCompile("明(?!年)", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			flag[2] = true
			cur = cur.AddDate(0, 0, 1)
		}
	}
	{
		pattern := regexp2.MustCompile("(?<!大)后天", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			flag[2] = true
			cur = cur.AddDate(0, 0, 2)
		}
	}
	{
		pattern := regexp.MustCompile(`大*大后天`)
		if pattern.MatchString(t.expTime) {
			flag[2] = true
			cnt := strings.Count(t.expTime, "大")
			cur = cur.AddDate(0, 0, -1*(2+cnt))
		}
	}
	// todo 补充星期相关的预测 done
	{
		pattern := regexp2.MustCompile(`(?<=(上*上上(周|星期)))[1-7]?`, 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			flag[2] = true
			week, err := strconv.Atoi(match.String())
			if err != nil {
				week = 1
			}
			if week == 7 {
				week = 0
			}
			cnt := strings.Count(t.expTime, "上")
			span := (week - int(cur.Weekday())) - 7*cnt
			cur = cur.AddDate(0, 0, span)
		}
	}
	{
		pattern := regexp2.MustCompile(`(?<=((?<!上)上(周|星期)))[1-7]?`, 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			flag[2] = true
			week, err := strconv.Atoi(match.String())
			if err != nil {
				week = 1
			}
			if week == 7 {
				week = 0
			}
			span := (week - int(cur.Weekday())) - 7
			cur = cur.AddDate(0, 0, span)
		}
	}
	{
		pattern := regexp2.MustCompile(`(?<=((?<!下)下(周|星期)))[1-7]?`, 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			flag[2] = true
			week, err := strconv.Atoi(match.String())
			if err != nil {
				week = 1
			}
			if week == 7 {
				week = 0
			}
			span := (week - int(cur.Weekday())) + 7
			cur = cur.AddDate(0, 0, span)
		}
	}
	// 这里对下下下周的时间转换做出了改善
	{
		pattern := regexp2.MustCompile(`(?<=(下*下下(周|星期)))[1-7]?`, 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			flag[2] = true
			week, err := strconv.Atoi(match.String())
			if err != nil {
				week = 1
			}
			if week == 7 {
				week = 0
			}
			cnt := strings.Count(t.expTime, "下")
			span := (week - int(cur.Weekday())) + 7*cnt
			cur = cur.AddDate(0, 0, span)
		}
	}
	{
		pattern := regexp2.MustCompile(`(?<=((?<!(上|下|个|[0-9]))(周|星期)))[1-7]`, 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			flag[2] = true
			week, err := strconv.Atoi(match.String())
			if err != nil {
				week = 1
			}
			if week == 7 {
				week = 0
			}
			span := week - int(cur.Weekday())
			cur = cur.AddDate(0, 0, span)
			// 处理未来时间
			cur = t.preferFutureWeek(week, cur)
		}
	}
	if flag[0] || flag[1] || flag[2] {
		t.tp[0] = cur.Year()
	}
	if flag[1] || flag[2] {
		t.tp[1] = int(cur.Month())
	}
	if flag[2] {
		t.tp[2] = cur.Day()
	}
}

// normSetHour 时-规范化方法：该方法识别时间表达式单元的时字段
func (t *TimeUnit) normSetHour() {
	pattern := regexp2.MustCompile("(?<!(周|星期))([0-2]?[0-9])(?=(点|时))", 0)
	if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
		h, _ := strconv.Atoi(match.String())
		t.tp[3] = h
		t.normCheckKeyword()
		t.preferFuture(3)
		t.isAllDayTime = true
	} else {
		t.normCheckKeyword()
	}
}

// normSetMinute 分-规范化方法：该方法识别时间表达式单元的分字段
func (t *TimeUnit) normSetMinute() {
	{
		pattern := regexp2.MustCompile("([0-9]+(?=分(?!钟)))|((?<=((?<!小)[点时]))[0-5]?[0-9](?!刻))", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			if minute, err := strconv.Atoi(match.String()); err == nil {
				t.tp[4] = minute
				t.isAllDayTime = false
			}
		}
	}
	{
		pattern := regexp2.MustCompile("(?<=[点时])[1一]刻(?!钟)", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			t.tp[4] = 15
			t.isAllDayTime = false
		}
	}
	{
		pattern := regexp2.MustCompile("(?<=[点时])半", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			t.tp[4] = 30
			t.isAllDayTime = false
		}
	}
	{
		pattern := regexp2.MustCompile("(?<=[点时])[3三]刻(?!钟)", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			t.tp[4] = 45
			t.isAllDayTime = false
		}
	}
}

// normSetSecond 添加了省略“秒”说法的时间：如17点15分32
func (t *TimeUnit) normSetSecond() {
	pattern := regexp2.MustCompile("([0-9]+(?=秒))|((?<=分)[0-5]?[0-9])", 0)
	if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
		sec, _ := strconv.Atoi(match.String())
		t.tp[5] = sec
		t.isAllDayTime = false
	}
}

// normSetSpecial 特殊形式的规范化方法-该方法识别特殊形式的时间表达式单元的各个字段
func (t *TimeUnit) normSetSpecial() {
	{
		pattern := regexp2.MustCompile("(晚上|夜间|夜里|今晚|明晚|晚|夜里|下午|午后)(?<!(周|星期))([0-2]?[0-9]):[0-5]?[0-9]:[0-5]?[0-9]", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			re := regexp.MustCompile(`([0-2]?[0-9]):[0-5]?[0-9]:[0-5]?[0-9]`)
			match := re.FindAllString(t.expTime, -1)
			for _, m := range match {
				parts := strings.Split(m, ":")
				if h, err := strconv.Atoi(parts[0]); err == nil {
					if h >= 0 && h <= 11 {
						t.tp[3] = h + 12
					} else {
						t.tp[3] = h
					}
				}
				if minute, err := strconv.Atoi(parts[1]); err == nil {
					t.tp[4] = minute
				}
				if sec, err := strconv.Atoi(parts[2]); err == nil {
					t.tp[5] = sec
				}
				t.preferFuture(3)
				t.isAllDayTime = false
				break
			}
			return
		}
	}
	{
		pattern := regexp2.MustCompile("(晚上|夜间|夜里|今晚|明晚|晚|夜里|下午|午后)(?<!(周|星期))([0-2]?[0-9]):[0-5]?[0-9]", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			re := regexp.MustCompile(`([0-2]?[0-9]):[0-5]?[0-9]`)
			match := re.FindAllString(t.expTime, -1)
			for _, m := range match {
				parts := strings.Split(m, ":")
				if h, err := strconv.Atoi(parts[0]); err == nil {
					if h >= 0 && h <= 11 {
						t.tp[3] = h + 12
					} else {
						t.tp[3] = h
					}
				}
				if minute, err := strconv.Atoi(parts[1]); err == nil {
					t.tp[4] = minute
				}
				t.preferFuture(3)
				t.isAllDayTime = false
				break
			}
			return
		}
	}
	{
		pattern := regexp2.MustCompile("(?<!(周|星期))([0-2]?[0-9]):[0-5]?[0-9]:[0-5]?[0-9](PM|pm|p\\.m)", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			re := regexp.MustCompile(`([0-2]?[0-9]):[0-5]?[0-9]:[0-5]?[0-9]`)
			match := re.FindAllString(t.expTime, -1)
			for _, m := range match {
				parts := strings.Split(m, ":")
				if h, err := strconv.Atoi(parts[0]); err == nil {
					if h >= 0 && h <= 11 {
						t.tp[3] = h + 12
					} else {
						t.tp[3] = h
					}
				}
				if minute, err := strconv.Atoi(parts[1]); err == nil {
					t.tp[4] = minute
				}
				if sec, err := strconv.Atoi(parts[2]); err == nil {
					t.tp[5] = sec
				}
				t.preferFuture(3)
				t.isAllDayTime = false
				break
			}
			return
		}
	}
	{
		pattern := regexp2.MustCompile("(?<!(周|星期))([0-2]?[0-9]):[0-5]?[0-9](PM|pm|p.m)", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			re := regexp.MustCompile(`([0-2]?[0-9]):[0-5]?[0-9]:[0-5]?[0-9]`)
			match := re.FindAllString(t.expTime, -1)
			for _, m := range match {
				parts := strings.Split(m, ":")
				if h, err := strconv.Atoi(parts[0]); err == nil {
					if h >= 0 && h <= 11 {
						t.tp[3] = h + 12
					} else {
						t.tp[3] = h
					}
				}
				if minute, err := strconv.Atoi(parts[1]); err == nil {
					t.tp[4] = minute
				}
				t.preferFuture(3)
				t.isAllDayTime = false
				break
			}
			return
		}
	}
	{
		pattern := regexp2.MustCompile("(?<!(周|星期|晚上|夜间|夜里|今晚|明晚|晚|夜里|下午|午后))([0-2]?[0-9]):[0-5]?[0-9]:[0-5]?[0-9]", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			parts := strings.Split(match.String(), ":")
			if h, err := strconv.Atoi(parts[0]); err == nil {
				t.tp[3] = h
			}
			if minute, err := strconv.Atoi(parts[1]); err == nil {
				t.tp[4] = minute
			}
			if sec, err := strconv.Atoi(parts[2]); err == nil {
				t.tp[5] = sec
			}
			t.preferFuture(3)
			t.isAllDayTime = false
			return
		}
	}
	{
		pattern := regexp2.MustCompile("(?<!(周|星期|晚上|夜间|夜里|今晚|明晚|晚|夜里|下午|午后))([0-2]?[0-9]):[0-5]?[0-9]", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			parts := strings.Split(match.String(), ":")
			if h, err := strconv.Atoi(parts[0]); err == nil {
				t.tp[3] = h
			}
			if minute, err := strconv.Atoi(parts[1]); err == nil {
				t.tp[4] = minute
			}
			t.preferFuture(3)
			t.isAllDayTime = false
			return
		}
	}
	// 这里是对年份表达的极好方式
	{
		pattern := regexp2.MustCompile("[0-9]?[0-9]?[0-9]{2}-((10)|(11)|(12)|([1-9]))-((?<!\\d))([0-3][0-9]|[1-9])", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			parts := strings.Split(match.String(), "-")
			if year, err := strconv.Atoi(parts[0]); err == nil {
				t.tp[0] = year
			}
			if month, err := strconv.Atoi(parts[1]); err == nil {
				t.tp[1] = month
			}
			if day, err := strconv.Atoi(parts[2]); err == nil {
				t.tp[2] = day
			}
		}
	}
	{
		pattern := regexp2.MustCompile("[0-9]?[0-9]?[0-9]{2}/((10)|(11)|(12)|([1-9]))/((?<!\\d))([0-3][0-9]|[1-9])", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			parts := strings.Split(match.String(), "/")
			if year, err := strconv.Atoi(parts[0]); err == nil {
				t.tp[0] = year
			}
			if month, err := strconv.Atoi(parts[1]); err == nil {
				t.tp[1] = month
			}
			if day, err := strconv.Atoi(parts[2]); err == nil {
				t.tp[2] = day
			}
		}
	}
	{
		pattern := regexp2.MustCompile("((10)|(11)|(12)|([1-9]))/((?<!\\d))([0-3][0-9]|[1-9])/[0-9]?[0-9]?[0-9]{2}", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			parts := strings.Split(match.String(), "/")
			if year, err := strconv.Atoi(parts[0]); err == nil {
				t.tp[0] = year
			}
			if month, err := strconv.Atoi(parts[1]); err == nil {
				t.tp[1] = month
			}
			if day, err := strconv.Atoi(parts[2]); err == nil {
				t.tp[2] = day
			}
		}
	}
	{
		pattern := regexp2.MustCompile("[0-9]?[0-9]?[0-9]{2}\\.((10)|(11)|(12)|([1-9]))\\.((?<!\\d))([0-3][0-9]|[1-9])", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			parts := strings.Split(match.String(), ".")
			if year, err := strconv.Atoi(parts[0]); err == nil {
				t.tp[0] = year
			}
			if month, err := strconv.Atoi(parts[1]); err == nil {
				t.tp[1] = month
			}
			if day, err := strconv.Atoi(parts[2]); err == nil {
				t.tp[2] = day
			}
		}
	}
}

// normSetSpanRelated 设置时间长度相关的时间表达式
func (t *TimeUnit) normSetSpanRelated() {
	{
		pattern := regexp2.MustCompile("\\d+(?=个月(?![以之]?[前后]))", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			t.normalizer.isTimeSpan = true
			month, _ := strconv.Atoi(match.String())
			t.tp[1] = month
		}
	}
	{
		pattern := regexp2.MustCompile("\\d+(?=天(?![以之]?[前后]))", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			t.normalizer.isTimeSpan = true
			day, _ := strconv.Atoi(match.String())
			t.tp[2] = day
		}
	}
	{
		pattern := regexp2.MustCompile("\\d+(?=(个)?小时(?![以之]?[前后]))", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			t.normalizer.isTimeSpan = true
			h, _ := strconv.Atoi(match.String())
			t.tp[3] = h
		}
	}
	{
		pattern := regexp2.MustCompile("\\d+(?=分钟(?![以之]?[前后]))", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			t.normalizer.isTimeSpan = true
			minute, _ := strconv.Atoi(match.String())
			t.tp[4] = minute
		}
	}
	{
		pattern := regexp2.MustCompile("\\d+(?=秒钟(?![以之]?[前后]))", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			t.normalizer.isTimeSpan = true
			second, _ := strconv.Atoi(match.String())
			t.tp[5] = second
		}
	}
	{
		pattern := regexp2.MustCompile("\\d+(?=(个)?(周|星期|礼拜)(?![以之]?[前后]))", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			t.normalizer.isTimeSpan = true
			week, _ := strconv.Atoi(match.String())
			if t.tp[2] == -1 {
				t.tp[2] = 0
			}
			t.tp[2] += week * 7
		}
	}
}

// normSetHoliday 节假日相关
func (t *TimeUnit) normSetHoliday() {
	pattern := regexp.MustCompile("(情人节)|(母亲节)|(青年节)|(教师节)|(中元节)|(端午)|(劳动节)|(7夕)|(建党节)|(建军节)|(初13)|(初14)|(初15)|(初12)|(初11)|(初9)|(初8)|(初7)|(初6)|(初5)|(初4)|(初3)|(初2)|(初1)|(中和节)|(圣诞)|(中秋)|(春节)|(元宵)|(航海日)|(儿童节)|(国庆)|(植树节)|(元旦)|(重阳节)|(妇女节)|(记者节)|(立春)|(雨水)|(惊蛰)|(春分)|(清明)|(谷雨)|(立夏)|(小满 )|(芒种)|(夏至)|(小暑)|(大暑)|(立秋)|(处暑)|(白露)|(秋分)|(寒露)|(霜降)|(立冬)|(小雪)|(大雪)|(冬至)|(小寒)|(大寒)")
	match := pattern.FindAllString(t.expTime, -1)
	for _, holi := range match {
		if t.tp[0] == -1 {
			t.tp[0] = t.normalizer.timeBase.Year()
		}
		if !strings.HasSuffix(holi, "节") {
			holi += "节"
		}
		date := make([]int, 2)
		if solar, found := t.normalizer.holiSolar[holi]; found {
			arr := strings.Split(solar, "-")
			date[0], _ = strconv.Atoi(arr[0])
			date[1], _ = strconv.Atoi(arr[1])
		} else if lunarDate, found := t.normalizer.holiLunar[holi]; found {
			arr := strings.Split(lunarDate, "-")
			date[0], _ = strconv.Atoi(arr[0])
			date[1], _ = strconv.Atoi(arr[1])
			lsConverter := NewLunarSolarConverter()
			lunar := Lunar{
				Year:  t.tp[0],
				Month: date[0],
				Day:   date[1],
			}
			solar := lsConverter.LunarToSolar(lunar)
			t.tp[0] = solar.Year
			date[0] = solar.Month
			date[1] = solar.Day
		} else {
			holi = strings.TrimSuffix(holi, "节")
			if holi == "小寒" || holi == "大寒" {
				t.tp[0] += 1
			}
			date = t.china24St(t.tp[0], holi)
		}
		t.tp[1] = date[0]
		t.tp[2] = date[1]
		break
	}
}

// normSetTotal 特殊形式的规范化方法
// 该方法识别特殊形式的时间表达式单元的各个字段
func (t *TimeUnit) normSetTotal() {
	{
		pattern := regexp2.MustCompile("(?<!(周|星期))([0-2]?[0-9]):[0-5]?[0-9]:[0-5]?[0-9]", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			arr := strings.Split(match.String(), ":")
			t.tp[3], _ = strconv.Atoi(arr[0])
			t.tp[4], _ = strconv.Atoi(arr[1])
			t.tp[5], _ = strconv.Atoi(arr[2])
			// 处理倾向于未来时间的情况
			t.preferFuture(3)
			t.isAllDayTime = false
		} else {
			pattern := regexp2.MustCompile("(?<!(周|星期))([0-2]?[0-9]):[0-5]?[0-9]", 0)
			if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
				arr := strings.Split(match.String(), ":")
				t.tp[3], _ = strconv.Atoi(arr[0])
				t.tp[4], _ = strconv.Atoi(arr[1])
				// 处理倾向于未来时间的情况
				t.preferFuture(3)
				t.isAllDayTime = false
			}
		}
	}
	// 增加了:固定形式时间表达式的
	// 中午,午间,下午,午后,晚上,傍晚,晚间,晚,pm,PM
	// 的正确时间计算，规约同上
	{
		pattern := regexp2.MustCompile("(中午)|(午间)", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			if t.tp[3] >= 0 && t.tp[3] <= 10 {
				t.tp[3] += 12
			} else if t.tp[3] == -1 {
				t.tp[3] = int(NOON)
			}
			// 处理倾向于未来时间的情况
			t.preferFuture(3)
			t.isAllDayTime = false
		}
	}
	{
		pattern := regexp2.MustCompile("(下午)|(午后)|(pm)|(PM)", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			if t.tp[3] >= 0 && t.tp[3] <= 11 {
				t.tp[3] += 12
			} else if t.tp[3] == -1 {
				t.tp[3] = int(AFTERNOON)
			}
			// 处理倾向于未来时间的情况
			t.preferFuture(3)
			t.isAllDayTime = false
		}
	}
	{
		pattern := regexp2.MustCompile("晚", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			if t.tp[3] >= 0 && t.tp[3] <= 11 {
				t.tp[3] += 12
			} else if t.tp[3] == 12 {
				t.tp[3] = 0
			} else if t.tp[3] == -1 {
				t.tp[3] = int(NIGHT)
			}
			// 处理倾向于未来时间的情况
			t.preferFuture(3)
			t.isAllDayTime = false
		}
	}
	{
		pattern := regexp2.MustCompile("[0-9]?[0-9]?[0-9]{2}-((10)|(11)|(12)|([1-9]))-((?<!\\d))([0-3][0-9]|[1-9])", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			arr := strings.Split(match.String(), "-")
			t.tp[0], _ = strconv.Atoi(arr[0])
			t.tp[1], _ = strconv.Atoi(arr[1])
			t.tp[2], _ = strconv.Atoi(arr[2])
		}
	}
	{
		pattern := regexp2.MustCompile("((10)|(11)|(12)|([1-9]))/((?<!\\d))([0-3][0-9]|[1-9])/[0-9]?[0-9]?[0-9]{2}", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			arr := strings.Split(match.String(), "/")
			t.tp[0], _ = strconv.Atoi(arr[0])
			t.tp[1], _ = strconv.Atoi(arr[1])
			t.tp[2], _ = strconv.Atoi(arr[2])
		}
	}
	// 增加了:固定形式时间表达式 年.月.日 的正确识别
	{
		pattern := regexp2.MustCompile("[0-9]?[0-9]?[0-9]{2}\\.((10)|(11)|(12)|([1-9]))\\.((?<!\\d))([0-3][0-9]|[1-9])", 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			arr := strings.Split(match.String(), "/")
			t.tp[0], _ = strconv.Atoi(arr[0])
			t.tp[1], _ = strconv.Atoi(arr[1])
			t.tp[2], _ = strconv.Atoi(arr[2])
		}
	}
}

// modifyTimeBase 该方法用于更新timeBase使之具有上下文关联性
func (t *TimeUnit) modifyTimeBase() {
	if !t.normalizer.isTimeSpan {
		if t.tp[0] >= 30 && t.tp[0] < 100 {
			t.tp[0] = 1900 + t.tp[0]
		} else if t.tp[0] > 0 && t.tp[0] < 30 {
			t.tp[0] = 2000 + t.tp[0]
		}
		timeGrid := NewTimePointFromTime(t.normalizer.timeBase)
		for idx, v := range t.tp {
			if v != -1 {
				timeGrid[idx] = t.tp[idx]
			}
		}
		t.normalizer.timeBase = timeGrid.ToTime(t.normalizer.timeBase.Location())
	}
}

// SolarTermData 阳历时间点数据
type SolorTermData struct {
	Key   float64
	Month int
	Years [][]int
}

// china24St 二十世纪和二十一世纪，24节气计算
// :param year: 年份
// :param china_st: 节气
//  :return: 节气日期（月, 日）
func (t *TimeUnit) china24St(year int, chinaSt string) []int {
	var stKey []float64
	if year/100 == 19 || year == 2000 {
		// 20世纪 key值
		stKey = []float64{6.11, 20.84, 4.6295, 19.4599, 6.3826, 21.4155, 5.59, 20.888, 6.318, 21.86, 6.5, 22.2, 7.928, 23.65, 8.35, 23.95, 8.44, 23.822, 9.098, 24.218, 8.218, 23.08, 7.9, 22.6}
	} else {
		// 21世纪 key值
		stKey = []float64{5.4055, 20.12, 3.87, 18.73, 5.63, 20.646, 4.81, 20.1, 5.52, 21.04, 5.678, 21.37, 7.108, 22.83, 7.5, 23.13, 7.646, 23.042, 8.318, 23.438, 7.438, 22.36, 7.18, 21.94}
	}
	// 二十四节气字典-- key值, 月份，(特殊年份，相差天数)...
	solorTerms := map[string]SolorTermData{
		"小寒": {
			Key:   stKey[0],
			Month: 1,
			Years: [][]int{{2019, -1}, {1982, 1}},
		},
		"大寒": {
			Key:   stKey[1],
			Month: 1,
			Years: [][]int{{2082, 1}},
		},
		"立春": {
			Key:   stKey[2],
			Month: 2,
			Years: [][]int{{-2, 0}},
		},
		"雨水": {
			Key:   stKey[3],
			Month: 2,
			Years: [][]int{{2026, -1}},
		},
		"惊蛰": {
			Key:   stKey[4],
			Month: 3,
			Years: [][]int{{-2, 0}},
		},
		"春分": {
			Key:   stKey[5],
			Month: 3,
			Years: [][]int{{2084, 1}},
		},
		"清明": {
			Key:   stKey[6],
			Month: 4,
			Years: [][]int{{-2, 0}},
		},
		"谷雨": {
			Key:   stKey[7],
			Month: 4,
			Years: [][]int{{-2, 0}},
		},
		"立夏": {
			Key:   stKey[8],
			Month: 5,
			Years: [][]int{{1911, 1}},
		},
		"小满": {
			Key:   stKey[9],
			Month: 5,
			Years: [][]int{{2008, 1}},
		},
		"芒种": {
			Key:   stKey[10],
			Month: 6,
			Years: [][]int{{1902, 1}},
		},
		"夏至": {
			Key:   stKey[11],
			Month: 6,
			Years: [][]int{{-2, 0}},
		},
		"小暑": {
			Key:   stKey[12],
			Month: 7,
			Years: [][]int{{2016, 1}, {1925, 1}},
		},
		"大暑": {
			Key:   stKey[13],
			Month: 7,
			Years: [][]int{{1922, 1}},
		},
		"立秋": {
			Key:   stKey[14],
			Month: 8,
			Years: [][]int{{2002, 1}},
		},
		"处暑": {
			Key:   stKey[15],
			Month: 8,
			Years: [][]int{{-2, 0}},
		},
		"白露": {
			Key:   stKey[16],
			Month: 9,
			Years: [][]int{{1927, 1}},
		},
		"秋分": {
			Key:   stKey[17],
			Month: 9,
			Years: [][]int{{-2, 0}},
		},
		"寒露": {
			Key:   stKey[18],
			Month: 10,
			Years: [][]int{{2088, 0}},
		},
		"霜降": {
			Key:   stKey[19],
			Month: 10,
			Years: [][]int{{2089, 1}},
		},
		"立冬": {
			Key:   stKey[20],
			Month: 11,
			Years: [][]int{{2089, 1}},
		},
		"小雪": {
			Key:   stKey[21],
			Month: 11,
			Years: [][]int{{1978, 0}},
		},
		"大雪": {
			Key:   stKey[22],
			Month: 12,
			Years: [][]int{{1954, 1}},
		},
		"冬至": {
			Key:   stKey[23],
			Month: 12,
			Years: [][]int{{2021, -1}, {1918, -1}},
		},
	}
	var flagDay int
	if chinaSt == "小寒" || chinaSt == "大寒" || chinaSt == "立春" || chinaSt == "雨水" {
		flagDay = int(float64(year%100)*0.2422+solorTerms[chinaSt].Key) - int((float64(year%100)-1)/4)
	} else {
		flagDay = int(float64(year%100)*0.2422+solorTerms[chinaSt].Key) - int(float64(year%100)/4)
	}
	// 特殊年份处理
	for _, spec := range solorTerms[chinaSt].Years {
		if spec[0] == year {
			flagDay += spec[1]
			break
		}
	}
	return []int{solorTerms[chinaSt].Month, flagDay}
}

// normCheckKeywor  对关键字：早（包含早上/早晨/早间），上午，中午,午间,下午,午后,晚上,傍晚,晚间,晚,pm,PM的正确时间计算
// 规约：
// 1. 中午/午间0-10点视为12-22点
// 2. 下午/午后0-11点视为12-23点
// 3. 晚上/傍晚/晚间/晚1-11点视为13-23点，12点视为0点
// 4. 0-11点pm/PM视为12-23点
func (t *TimeUnit) normCheckKeyword() {
	if strings.Contains(t.expTime, "凌晨") {
		t.isMorning = true
		if t.tp[3] == -1 {
			// 增加对没有明确时间点，只写了“凌晨”这种情况的处理
			t.tp[3] = int(DAY_BREAK)
		} else if t.tp[3] > 12 && t.tp[3] <= 23 {
			t.tp[3] -= 12
		} else if t.tp[3] == 0 {
			t.tp[3] = 12
		}
		// 处理倾向于未来时间的情况
		t.preferFuture(3)
		t.isAllDayTime = false
	}
	{
		pattern := regexp.MustCompile(`早上|早晨|早间|晨间|今早|明早|早|清晨`)
		if pattern.MatchString(t.expTime) {
			t.isMorning = true
			if t.tp[3] == -1 {
				// 增加对没有明确时间点，只写了“早上/早晨/早间”这种情况的处理
				t.tp[3] = int(EARLY_MORNING)
			} else if t.tp[3] > 12 && t.tp[3] <= 23 {
				t.tp[3] -= 12
			} else if t.tp[3] == 0 {
				t.tp[3] = 12
			}
			t.preferFuture(3)
			t.isAllDayTime = false
		}
	}
	if strings.Contains(t.expTime, "上午") {
		t.isMorning = true
		if t.tp[3] == -1 {
			// 增加对没有明确时间点，只写了“上午”这种情况的处理
			t.tp[3] = int(MORNING)
		} else if t.tp[3] > 12 && t.tp[3] <= 23 {
			t.tp[3] -= 12
		} else if t.tp[3] == 0 {
			t.tp[3] = 12
		}
		// 处理倾向于未来时间的情况
		t.preferFuture(3)
		t.isAllDayTime = false
	}
	{
		pattern := regexp.MustCompile(`(中午)|(午间)|白天`)
		if pattern.MatchString(t.expTime) {
			t.isMorning = true
			if t.tp[3] >= 0 && t.tp[3] <= 10 {
				t.tp[3] += 12
			} else if t.tp[3] == -1 {
				// 增加对没有明确时间点，只写了“中午/午间”这种情况的处理
				t.tp[3] = int(NOON)
			}
			t.preferFuture(3)
			t.isAllDayTime = false
		}
	}
	{
		pattern := regexp.MustCompile(`(下午)|(午后)|(pm)|(PM)`)
		if pattern.MatchString(t.expTime) {
			if t.tp[3] >= 0 && t.tp[3] <= 11 {
				t.tp[3] += 12
			} else if t.tp[3] == -1 {
				// 增加对没有明确时间点，只写了“中午/午间”这种情况的处理
				t.tp[3] = int(AFTERNOON)
			}
			t.preferFuture(3)
			t.isAllDayTime = false
		}
	}
	{
		pattern := regexp.MustCompile(`晚上|夜间|夜里|今晚|明晚|晚|夜里`)
		if pattern.MatchString(t.expTime) {
			if t.tp[3] >= 0 && t.tp[3] <= 11 {
				t.tp[3] += 12
			} else if t.tp[3] == 12 {
				t.tp[3] = 0
			} else if t.tp[3] == -1 {
				// 增加对没有明确时间点，只写了“中午/午间”这种情况的处理
				t.tp[3] = int(LATE_NIGHT)
			}
			t.preferFuture(3)
			t.isAllDayTime = false
		}
	}
}

// preferFutureWeek 未来星期几
func (t *TimeUnit) preferFutureWeek(week int, cur time.Time) time.Time {
	// 确认用户选项
	if !t.normalizer.isPreferFuture {
		return cur
	}
	// 检查被检查的时间级别之前，是否没有更高级的已经确定的时间，如果有，则不进行处理.
	if t.tp[0] != -1 || t.tp[1] != -1 || t.tp[2] != -1 {
		return cur
	}
	// 获取当前是在周几，如果识别到的时间小于当前时间，则识别时间为下一周
	if int(time.Now().Weekday()) > week {
		cur = cur.AddDate(0, 0, 7)
	}
	return cur
}

// preferFuture 如果用户选项是倾向于未来时间，检查checkTimeIndex所指的时间是否是过去的时间，如果是的话，将大一级的时间设为当前时间的+1。
// 如在晚上说“早上8点看书”，则识别为明天早上;
// 12月31日说“3号买菜”，则识别为明年1月的3号。
// :param checkTimeIndex: _tp.tunit时间数组的下标
func (t *TimeUnit) preferFuture(checkTimeIndex int) {
	// 1. 检查被检查的时间级别之前，是否没有更高级的已经确定的时间，如果有，则不进行处理.
	{
		var idx = 0
		for idx < checkTimeIndex {
			if t.tp[idx] != -1 {
				return
			}
			idx += 1
		}
	}
	// 2. 根据上下文补充时间
	t.checkContextTime(checkTimeIndex)
	// 3. 根据上下文补充时间后再次检查被检查的时间级别之前，是否没有更高级的已经确定的时间，如果有，则不进行倾向处理.
	/*
		{
			var idx = 0
			for idx < checkTimeIndex {
				if t.tp[idx] != -1 {
					return
				}
				idx += 1
			}
		}
	*/
	// 4. 确认用户选项
	if !t.normalizer.isPreferFuture {
		return
	}
	// 5. 获取当前时间，如果识别到的时间小于当前时间，则将其上的所有级别时间设置为当前时间，并且其上一级的时间步长+1
	basePoint := NewTimePointFromTime(t.normalizer.timeBase)
	t.noYear = t.tp[0] == -1
	if basePoint[checkTimeIndex] < t.tp[checkTimeIndex] {
		return
	}
	// 准备增加的时间单位是被检查的时间的上一级，将上一级时间+1
	{
		curr := t.addTime(t.normalizer.timeBase, checkTimeIndex-1)
		currPoint := NewTimePointFromTime(curr)
		var idx = 0
		for idx < checkTimeIndex {
			t.tp[idx] = currPoint[idx]
			idx += 1
		}
	}
}

// checkTime 检查未来时间点
func (t *TimeUnit) checkTime(timePoint TimePoint) {
	if !t.noYear {
		return
	}
	// check the month
	basePoint := NewTimePointFromTime(t.normalizer.timeBase)
	if timePoint[1] == basePoint[1] && timePoint[2] > basePoint[2] {
		timePoint[0] -= 1
	}
	t.noYear = false
}

// checkContextTime 根据上下文时间补充时间信息
func (t *TimeUnit) checkContextTime(checkTimeIndex int) {
	if !t.isFirstTimeSolveContext {
		return
	}
	var idx = 0
	for idx < checkTimeIndex {
		if t.tp[idx] == -1 && t.tpOrigin[idx] != -1 {
			t.tp[idx] = t.tpOrigin[idx]
		}
		idx += 1
	}
	to := t.tpOrigin[checkTimeIndex]
	tp := t.tp[checkTimeIndex]
	if !t.isMorning && checkTimeIndex == 3 && to >= 12 && (to-12) < tp && tp < 12 {
		t.tp[checkTimeIndex] += 12
	}
	t.isFirstTimeSolveContext = false
}

func (t *TimeUnit) addTime(baseTime time.Time, foreUnit int) time.Time {
	switch foreUnit {
	case 0:
		return baseTime.AddDate(1, 0, 0)
	case 1:
		return baseTime.AddDate(0, 1, 0)
	case 2:
		return baseTime.AddDate(0, 0, 1)
	case 3:
		return baseTime.Add(time.Hour * 1)
	case 4:
		return baseTime.Add(time.Minute * 1)
	case 5:
		return baseTime.Add(time.Second * 1)
	}
	return baseTime
}
