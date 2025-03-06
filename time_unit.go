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
		t.normalizeTimeSpan()
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

func (t *TimeUnit) normalizeTimeSpan() {
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
	idx := 3
	for idx < 6 {
		if t.tp[idx] < 0 {
			tunit[idx] = 0
		} else {
			tunit[idx] = t.tp[idx]
		}
		idx++
	}
	seconds := int64(tunit[3]*3600) + int64(tunit[4]*60) + int64(tunit[5])
	if seconds == 0 && days == 0 {
		t.normalizer.isTimeSpan = false
		t.normalizer.invalidSpan = true
		return
	}
	t.ts = t.normalizer.timeBase.Add(t.genSpan(days, seconds)).Truncate(time.Second)
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
	pattern2 := regexp.MustCompile(`第(\d+)季`)
	if match := pattern2.FindAllStringSubmatch(t.expTime, 1); len(match) > 0 && len(match[0]) == 2 {
		if season, _ := strconv.Atoi(match[0][1]); season > 0 {
			cur := t.tp.ToTime(t.normalizer.timeBase.Location())
			cur = cur.AddDate(0, (season-1)*4, 0)
			t.tp[0] = cur.Year()
			t.tp[1] = int(cur.Month())
			t.preferFuture(1)
		}
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
	pattern2 := regexp.MustCompile(`第(\d+)周`)
	if match := pattern2.FindAllStringSubmatch(t.expTime, 1); len(match) > 0 && len(match[0]) == 2 {
		if weeks, _ := strconv.Atoi(match[0][1]); weeks > 0 {
			cur := t.tp.ToTime(t.normalizer.timeBase.Location())
			cur = cur.AddDate(0, 0, (weeks-1)*7)
			t.tp[0] = cur.Year()
			t.tp[1] = int(cur.Month())
			t.tp[2] = cur.Day()
		}
	}
}

// normSetMonthFuzzyDay 月-日 兼容模糊写法：该方法识别时间表达式单元的月、日字段
func (t *TimeUnit) normSetMonthFuzzyDay() {
	pattern := regexp.MustCompile(`((10)|(11)|(12)|([1-9]))(月|\.|\-)([0-3][0-9]|[1-9])`)
	match := pattern.FindAllString(t.expTime, -1)
	for _, m := range match {
		p := regexp.MustCompile(`(月|\.|\-)`)
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
	settings := []struct {
		Reg     string
		FlagIdx int
		Negtive bool
	}{
		{Reg: "\\d+(?=(个)?小时[以之]?前)", FlagIdx: 3, Negtive: true},
		{Reg: "\\d+(?=(个)?小时[以之]?后)", FlagIdx: 3},
		{Reg: "\\d+(?=天[以之]?前)", FlagIdx: 2, Negtive: true},
		{Reg: "\\d+(?=天[以之]?后)", FlagIdx: 2},
		{Reg: "\\d+(?=(个)?月[以之]?前)", FlagIdx: 1, Negtive: true},
		{Reg: "\\d+(?=(个)?月[以之]?后)", FlagIdx: 1},
		{Reg: "\\d+(?=年[以之]?前)", FlagIdx: 0, Negtive: true},
		{Reg: "\\d+(?=年[以之]?后)", FlagIdx: 0},
	}
	var updateFlag bool
	for _, s := range settings {
		cur, updateFlag = t.calcNormSetBaseRelated(cur, s.Reg, s.FlagIdx, s.Negtive)
		if updateFlag {
			flag[s.FlagIdx] = true
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

func (t *TimeUnit) calcNormSetBaseRelated(cur time.Time, reg string, flagIdx int, negtive bool) (time.Time, bool) {
	var update bool
	pattern := regexp2.MustCompile(reg, 0)
	if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
		update = true
		delta, _ := strconv.Atoi(match.String())
		if negtive {
			delta *= -1
		}
		switch flagIdx {
		case 3:
			cur = cur.Add(time.Duration(delta) * time.Hour)
		case 2:
			cur = cur.AddDate(0, 0, delta)
		case 1:
			cur = cur.AddDate(0, delta, 0)
		case 0:
			cur = cur.AddDate(delta, 0, 0)
		}
	}
	return cur, update
}

// normSetCurRelated 设置当前时间相关的时间表达式
func (t *TimeUnit) normSetCurRelated() {
	cur := time.Now()
	flag := []bool{false, false, false}
	var updateFlag bool
	cur, updateFlag = t.normSetCurRelatedYear(cur)
	if updateFlag {
		flag[0] = true
	}
	cur, updateFlag = t.normSetCurRelatedMonth(cur)
	if updateFlag {
		flag[1] = true
	}
	cur, updateFlag = t.normSetCurRelatedDay(cur)
	if updateFlag {
		flag[2] = true
	}
	cur, updateFlag = t.normSetCurRelatedWeek(cur)
	if updateFlag {
		flag[2] = true
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

func (t *TimeUnit) normSetCurRelatedYear(cur time.Time) (time.Time, bool) {
	var updateFlag bool
	yearSettings := []struct {
		Word  string
		Years int
	}{
		{Word: "前年", Years: -2},
		{Word: "去年", Years: -1},
		{Word: "今年", Years: 0},
		{Word: "明年", Years: 1},
		{Word: "后年", Years: 2},
	}
	for _, s := range yearSettings {
		var matched bool
		cur, matched = t.calcNormSetCurRelatedYear(cur, s.Word, s.Years)
		if matched {
			updateFlag = true
		}
	}
	return cur, updateFlag
}

func (t *TimeUnit) calcNormSetCurRelatedYear(cur time.Time, word string, years int) (time.Time, bool) {
	if strings.Contains(t.expTime, word) {
		if years != 0 {
			cur = cur.AddDate(years, 0, 0)
		}
		return cur, true
	}
	return cur, false
}

func (t *TimeUnit) calcNormSetCurRelatedMonth(cur time.Time, reg string, char string, negtive bool) (time.Time, bool) {
	pattern := regexp.MustCompile(reg)
	if pattern.MatchString(t.expTime) {
		if char != "" {
			cnt := strings.Count(t.expTime, char)
			if negtive {
				cnt *= -1
			}
			cur = cur.AddDate(0, cnt, 0)
		}
		return cur, true
	}
	return cur, false
}

func (t *TimeUnit) normSetCurRelatedMonth(cur time.Time) (time.Time, bool) {
	var updateFlag bool
	monthSettings := []struct {
		Reg     string
		Char    string
		Negtive bool
	}{
		{Reg: `上*上(个)?月`, Char: "上", Negtive: true},
		{Reg: `(本|这个)月`},
		{Reg: `下*下(个)?月`, Char: "下"},
	}
	for _, s := range monthSettings {
		var matched bool
		cur, matched = t.calcNormSetCurRelatedMonth(cur, s.Reg, s.Char, s.Negtive)
		if matched {
			updateFlag = matched
		}
	}
	return cur, updateFlag
}

func (t *TimeUnit) normSetCurRelatedDay(cur time.Time) (time.Time, bool) {
	var updateFlag bool
	daysSettings := []struct {
		Reg  string
		Char string
		Days int
	}{
		{Reg: `大*大前天`, Char: "大", Days: 0},
		{Reg: "(?<!大)前天", Days: -2},
		{Reg: "(?<!大)前天", Days: -1},
		{Char: "昨", Days: -1},
		{Reg: "今(?!年)"},
		{Reg: "明(?!年)", Days: 1},
		{Reg: "(?<!大)后天", Days: 2},
		{Reg: `大*大后天`, Char: "大", Days: 2},
	}
	for _, s := range daysSettings {
		var matched bool
		cur, matched = t.calcNormSetCurRelatedDay(cur, s.Reg, s.Char, s.Days)
		if matched {
			updateFlag = true
		}
	}
	return cur, updateFlag
}

func (t *TimeUnit) calcNormSetCurRelatedDay(cur time.Time, reg string, char string, days int) (time.Time, bool) {
	if reg == "" {
		if strings.Contains(t.expTime, char) {
			cur = cur.AddDate(0, 0, days)
			return cur, true
		}
		return cur, false
	}
	if char != "" {
		pattern := regexp.MustCompile(reg)
		if pattern.MatchString(t.expTime) {
			cnt := strings.Count(t.expTime, char)
			cur = cur.AddDate(0, 0, -1*(days+cnt))
			return cur, true
		}
		return cur, false
	}
	pattern := regexp2.MustCompile(reg, 0)
	if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
		if days != 0 {
			cur = cur.AddDate(0, 0, days)
		}
		return cur, true
	}
	return cur, false
}

func (t *TimeUnit) normSetCurRelatedWeek(cur time.Time) (time.Time, bool) {
	var updateFlag bool
	weekSettings := []struct {
		Reg          string
		Char         string
		Days         int
		PreferFuture bool
	}{
		{Reg: `(?<=(上*上上(周|星期)))[1-7]?`, Char: "上", Days: -7},
		{Reg: `(?<=((?<!上)上(周|星期)))[1-7]?`, Days: -7},
		{Reg: `(?<=((?<!下)下(周|星期)))[1-7]?`, Days: 7},
		{Reg: `(?<=(下*下下(周|星期)))[1-7]?`, Char: "下", Days: 7}, // 这里对下下下周的时间转换做出了改善
		{Reg: `(?<=((?<!(上|下|个|[0-9]))(周|星期)))[1-7]`, Days: 0, PreferFuture: true},
	}
	for _, s := range weekSettings {
		var matched bool
		cur, matched = t.calcNormSetCurRelatedWeek(cur, s.Reg, s.Char, s.Days, s.PreferFuture)
		if matched {
			updateFlag = true
		}
	}
	return cur, updateFlag
}

func (t *TimeUnit) calcNormSetCurRelatedWeek(cur time.Time, reg string, char string, days int, preferFuture bool) (time.Time, bool) {
	pattern := regexp2.MustCompile(reg, 0)
	if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
		week, err := strconv.Atoi(match.String())
		if err != nil {
			week = 1
		}
		if week == 7 {
			week = 0
		}
		var cnt int = 1
		if char != "" {
			cnt = strings.Count(t.expTime, char)
		}
		span := (week - int(cur.Weekday())) + days*cnt
		cur = cur.AddDate(0, 0, span)
		if preferFuture {
			// 处理未来时间
			cur = t.preferFutureWeek(week, cur)
		}
		return cur, true
	}
	return cur, false
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
	cases := []struct {
		Regs   []string
		HasSec bool
	}{
		{
			Regs: []string{
				"(晚上|夜间|夜里|今晚|明晚|晚|夜里|下午|午后)(?<!(周|星期))([0-2]?[0-9]):[0-5]?[0-9]:[0-5]?[0-9]",
				`([0-2]?[0-9]):[0-5]?[0-9]:[0-5]?[0-9]`,
			},
			HasSec: true,
		},
		{
			Regs: []string{
				"(晚上|夜间|夜里|今晚|明晚|晚|夜里|下午|午后)(?<!(周|星期))([0-2]?[0-9]):[0-5]?[0-9]",
				`([0-2]?[0-9]):[0-5]?[0-9]`,
			},
			HasSec: false,
		},
		{
			Regs: []string{
				"(?<!(周|星期))([0-2]?[0-9]):[0-5]?[0-9]:[0-5]?[0-9](PM|pm|p\\.m)",
				`([0-2]?[0-9]):[0-5]?[0-9]:[0-5]?[0-9]`,
			},
			HasSec: true,
		},
		{
			Regs: []string{
				"(?<!(周|星期))([0-2]?[0-9]):[0-5]?[0-9](PM|pm|p.m)",
				`([0-2]?[0-9]):[0-5]?[0-9]:[0-5]?[0-9]`,
			},
			HasSec: false,
		},
		{
			Regs:   []string{"(?<!(周|星期|晚上|夜间|夜里|今晚|明晚|晚|夜里|下午|午后))([0-2]?[0-9]):[0-5]?[0-9]:[0-5]?[0-9]"},
			HasSec: true,
		},
		{
			Regs:   []string{"(?<!(周|星期|晚上|夜间|夜里|今晚|明晚|晚|夜里|下午|午后))([0-2]?[0-9]):[0-5]?[0-9]"},
			HasSec: false,
		},
	}
	for _, c := range cases {
		if matched := t.calcNormSetSpecial(c.Regs, c.HasSec); matched {
			return
		}
	}
	yearSettings := []struct {
		Reg     string
		Spliter string
	}{
		{Reg: "[0-9]?[0-9]?[0-9]{2}-((10)|(11)|(12)|([1-9]))-((?<!\\d))([0-3][0-9]|[1-9])", Spliter: "-"},
		{Reg: "[0-9]?[0-9]?[0-9]{2}/((10)|(11)|(12)|([1-9]))/((?<!\\d))([0-3][0-9]|[1-9])", Spliter: "/"},
		{Reg: "[0-9]?[0-9]?[0-9]{2}\\.((10)|(11)|(12)|([1-9]))\\.((?<!\\d))([0-3][0-9]|[1-9])", Spliter: "."},
	}
	for _, s := range yearSettings {
		t.calcNormSetSpecialYear(s.Reg, s.Spliter)
	}
}

func (t *TimeUnit) calcNormSetSpecial(regs []string, hasSec bool) bool {
	pattern := regexp2.MustCompile(regs[0], 0)
	if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
		if len(regs) == 1 {
			parts := strings.Split(match.String(), ":")
			if h, err := strconv.Atoi(parts[0]); err == nil {
				t.tp[3] = h
			}
			if minute, err := strconv.Atoi(parts[1]); err == nil {
				t.tp[4] = minute
			}
			if hasSec {
				if sec, err := strconv.Atoi(parts[2]); err == nil {
					t.tp[5] = sec
				}
			}
			t.preferFuture(3)
			t.isAllDayTime = false
			return true
		}
		re := regexp.MustCompile(regs[1])
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
			if hasSec {
				if sec, err := strconv.Atoi(parts[2]); err == nil {
					t.tp[5] = sec
				}
			}
			t.preferFuture(3)
			t.isAllDayTime = false
			break
		}
		return true
	}
	return false
}

func (t *TimeUnit) calcNormSetSpecialYear(reg string, spliter string) {
	pattern := regexp2.MustCompile(reg, 0)
	if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
		parts := strings.Split(match.String(), spliter)
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

// normSetSpanRelated 设置时间长度相关的时间表达式
func (t *TimeUnit) normSetSpanRelated() {
	cases := []struct {
		Reg     string
		Idx     int
		AddWeek bool
	}{
		{Reg: "\\d+(?=个月(?![以之]?[前后]))", Idx: 1},
		{Reg: "\\d+(?=天(?![以之]?[前后]))", Idx: 2},
		{Reg: "\\d+(?=(个)?小时(?![以之]?[前后]))", Idx: 3},
		{Reg: `\d+(?=分钟(?![以之]?[前后]))`, Idx: 4},
		{Reg: `\d+(?=秒钟(?![以之]?[前后]))`, Idx: 5},
		{Reg: `(?<!第)\d+(?=(个)?(周|星期|礼拜)(?![以之]?[前后]))`, Idx: 2, AddWeek: true},
	}
	for _, c := range cases {
		pattern := regexp2.MustCompile(c.Reg, 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			t.normalizer.isTimeSpan = true
			value, _ := strconv.Atoi(match.String())
			if c.AddWeek {
				if t.tp[2] == -1 {
					t.tp[2] = 0
				}
				t.tp[2] += value * 7
			} else {
				t.tp[c.Idx] = value
			}
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
	t.calcNormSetTotalTime()
	t.calcNormSetTotalDaytime()
	t.calcNormSetTotalDay()
}

func (t *TimeUnit) calcNormSetTotalTime() {
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

// calcNormSetTotalDaytime 增加了:固定形式时间表达式的
// 中午,午间,下午,午后,晚上,傍晚,晚间,晚,pm,PM
// 的正确时间计算，规约同上
func (t *TimeUnit) calcNormSetTotalDaytime() {
	cases := []struct {
		Reg   string
		Point RangeTimeEnum
	}{
		{Reg: "(中午)|(午间)", Point: NOON},
		{Reg: "(下午)|(午后)|(pm)|(PM)", Point: AFTERNOON},
		{Reg: "晚", Point: NIGHT},
	}
	for _, c := range cases {
		endTime := 11
		if c.Point == NOON {
			endTime = 10
		}
		pattern := regexp2.MustCompile(c.Reg, 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			if t.tp[3] >= 0 && t.tp[3] <= endTime {
				t.tp[3] += 12
			} else if c.Point == NIGHT && t.tp[3] == 12 {
				t.tp[3] = 0
			} else if t.tp[3] == -1 {
				t.tp[3] = int(c.Point)
			}
			// 处理倾向于未来时间的情况
			t.preferFuture(3)
			t.isAllDayTime = false
		}
	}
}

func (t *TimeUnit) calcNormSetTotalDay() {
	cases := [][]string{
		{"[0-9]?[0-9]?[0-9]{2}-((10)|(11)|(12)|([1-9]))-((?<!\\d))([0-3][0-9]|[1-9])", "-"},
		{"((10)|(11)|(12)|([1-9]))/((?<!\\d))([0-3][0-9]|[1-9])/[0-9]?[0-9]?[0-9]{2}", "/"},
		{"[0-9]?[0-9]?[0-9]{2}\\.((10)|(11)|(12)|([1-9]))\\.((?<!\\d))([0-3][0-9]|[1-9])", "/"}, // 增加了:固定形式时间表达式 年.月.日 的正确识别
	}
	for _, c := range cases {
		pattern := regexp2.MustCompile(c[0], 0)
		if match, _ := pattern.FindStringMatch(t.expTime); match != nil {
			arr := strings.Split(match.String(), c[1])
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
type SolarTermData struct {
	// Key 索引值
	Key float64
	// Month 月份
	Month int
	// Years 年份
	Years [][]int
}

// china24St 二十世纪和二十一世纪，24节气计算
// :param year: 年份
// :param china_st: 节气
//
//	:return: 节气日期（月, 日）
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
	solarTerms := map[string]SolarTermData{
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
		flagDay = int(float64(year%100)*0.2422+solarTerms[chinaSt].Key) - int((float64(year%100)-1)/4)
	} else {
		flagDay = int(float64(year%100)*0.2422+solarTerms[chinaSt].Key) - int(float64(year%100)/4)
	}
	// 特殊年份处理
	for _, spec := range solarTerms[chinaSt].Years {
		if spec[0] == year {
			flagDay += spec[1]
			break
		}
	}
	return []int{solarTerms[chinaSt].Month, flagDay}
}

// normCheckKeyword  对关键字：早（包含早上/早晨/早间），上午，中午,午间,下午,午后,晚上,傍晚,晚间,晚,pm,PM的正确时间计算
// 规约：
// 1. 中午/午间0-10点视为12-22点
// 2. 下午/午后0-11点视为12-23点
// 3. 晚上/傍晚/晚间/晚1-11点视为13-23点，12点视为0点
// 4. 0-11点pm/PM视为12-23点
func (t *TimeUnit) normCheckKeyword() {
	cases := []struct {
		Reg   string
		Point RangeTimeEnum
	}{
		{Reg: "凌晨", Point: DAY_BREAK},
		{Reg: `早上|早晨|早间|晨间|今早|明早|早|清晨`, Point: EARLY_MORNING},
		{Reg: "上午", Point: MORNING},
		{Reg: `(中午)|(午间)|白天`, Point: NOON},
		{Reg: `(下午)|(午后)|(pm)|(PM)`, Point: AFTERNOON},
		{Reg: `晚上|夜间|夜里|今晚|明晚|晚|夜里`, Point: LATE_NIGHT},
	}
	for _, c := range cases {
		t.calcNormCheckKeyword(c.Reg, c.Point)
	}
}

func (t *TimeUnit) calcNormCheckKeyword(reg string, timepoint RangeTimeEnum) {
	var matched bool
	if timepoint == DAY_BREAK || timepoint == MORNING {
		matched = strings.Contains(t.expTime, reg)
	} else {
		pattern := regexp.MustCompile(reg)
		matched = pattern.MatchString(t.expTime)
	}
	if !matched {
		return
	}
	if timepoint == DAY_BREAK || timepoint == EARLY_MORNING || timepoint == MORNING {
		t.isMorning = true
		if t.tp[3] == -1 {
			// 增加对没有明确时间点，只写了“凌晨”这种情况的处理
			t.tp[3] = int(timepoint)
		} else if t.tp[3] > 12 && t.tp[3] <= 23 {
			t.tp[3] -= 12
		} else if t.tp[3] == 0 {
			t.tp[3] = 12
		}
	} else if timepoint == NOON {
		t.isMorning = true
		if t.tp[3] >= 0 && t.tp[3] <= 10 {
			t.tp[3] += 12
		} else if t.tp[3] == -1 {
			// 增加对没有明确时间点，只写了“中午/午间”这种情况的处理
			t.tp[3] = int(NOON)
		}
	} else {
		pattern := regexp.MustCompile(reg)
		matched = pattern.MatchString(t.expTime)
		if !matched {
			return
		}
		if t.tp[3] >= 0 && t.tp[3] <= 11 {
			t.tp[3] += 12
		} else if timepoint == LATE_NIGHT && t.tp[3] == 12 {
			t.tp[3] = 0
		} else if t.tp[3] == -1 {
			// 增加对没有明确时间点，只写了“中午/午间”这种情况的处理
			t.tp[3] = int(timepoint)
		}
	}
	t.preferFuture(3)
	t.isAllDayTime = false
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
		idx := 0
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
		idx := 0
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
	idx := 0
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
