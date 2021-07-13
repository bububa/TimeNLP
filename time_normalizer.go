package timenlp

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dlclark/regexp2"

	"github.com/bububa/TimeNLP/resource"
)

// TimeNormalizer 时间表达式识别的主要工作类
type TimeNormalizer struct {
	isPreferFuture bool
	isTimeSpan     bool
	invalidSpan    bool
	timeBase       time.Time
	pattern        *regexp2.Regexp
	holiSolar      map[string]string
	holiLunar      map[string]string
}

// NewTimeNormalizer 新建TimeNormalizer
// isPreferFuture: 是否倾向使用未来时间
func NewTimeNormalizer(isPreferFuture bool) *TimeNormalizer {
	pattern := regexp2.MustCompile(resource.Pattern, 0)
	holiSolar := make(map[string]string)
	holiLunar := make(map[string]string)
	json.Unmarshal(resource.HoliSolar, &holiSolar)
	json.Unmarshal(resource.HoliLunar, &holiLunar)
	return &TimeNormalizer{
		isPreferFuture: isPreferFuture,
		pattern:        pattern,
		holiSolar:      holiSolar,
		holiLunar:      holiLunar,
	}
}

// filter 这里对一些不规范的表达做转换
func (n *TimeNormalizer) filter(inputQuery string) string {
	preHandler := &StringPreHandler{}
	inputQuery = preHandler.NumberTranslator(inputQuery)
	{
		pattern := regexp.MustCompile("[0-9]月[0-9]")
		if pattern.MatchString(inputQuery) {
			index := strings.Index(inputQuery, "月")
			pattern := regexp.MustCompile("日|号")
			if !pattern.MatchString(inputQuery[index:]) {
				pattern := regexp.MustCompile("[0-9]月[0-9]+")
				if loc := pattern.FindStringIndex(inputQuery); loc != nil {
					inputQuery = fmt.Sprintf("%s号%s", inputQuery[0:loc[1]], inputQuery[loc[1]:])
				}
			}
		}
	}
	if !strings.Contains(inputQuery, "月") && !strings.Contains(inputQuery, "个半") {
		inputQuery = strings.Replace(inputQuery, "个", "", -1)
	}
	replaces := [][]string{
		{"中旬", "15号"},
		{"傍晚", "午后"},
		{"大年", ""},
		{"五一", "劳动节"},
		{"白天", "早上"},
		{"：", ":"},
	}
	for _, rpl := range replaces {
		inputQuery = strings.Replace(inputQuery, rpl[0], rpl[1], -1)
	}
	return inputQuery
}

// preHandling 待匹配字符串的清理空白符和语气助词以及大写数字转化的预处理
func (n *TimeNormalizer) preHandling(target string) string {
	preHandler := &StringPreHandler{}
	rules := []string{
		"\\s+",
		"[的]+",
	}
	for _, rule := range rules {
		target = preHandler.DelKeyword(target, rule)
	}
	target = preHandler.NumberTranslator(target)
	return target
}

// Parse 是TimeNormalizer的构造方法，根据提供的待分析字符串和timeBase进行时间表达式提取
func (n *TimeNormalizer) Parse(target string, timeBase time.Time) (*Result, error) {
	n.timeBase = timeBase
	target = n.preHandling(target)
	timeUnits := n.timeExt(target, timeBase)
	ret := Result{
		NormalizedString: target,
	}
	if len(timeUnits) == 0 {
		return nil, errors.New("no time pattern could be extracted.")
	} else if n.isTimeSpan && !n.invalidSpan {
		ret.Type = DELTA
	} else if len(timeUnits) == 1 {
		ret.Type = TIMESTAMP
	} else {
		ret.Type = SPAN
	}
	for _, v := range timeUnits {
		ret.Points = append(ret.Points, v.ToResultPoint())
	}
	return &ret, nil
}

// timeExt 有基准时间输入的时间表达式识别
// 这是时间表达式识别的主方法， 通过已经构建的正则表达式对字符串进行识别，并按照预先定义的基准时间进行规范化
// 将所有别识别并进行规范化的时间表达式进行返回， 时间表达式通过TimeUnit类进行定义
func (n *TimeNormalizer) timeExt(target string, timeBase time.Time) []TimeUnit {
	var (
		startLine int = -1
		endLine   int = -1
		rPointer  int = 0 // 计数器，记录当前识别到哪一个字符串了
		temp      []string
		pos       []int
		length    []int
	)
	m, _ := n.pattern.FindStringMatch(target)
	for m != nil {
		startLine = m.Index
		if startLine == endLine { // 假如下一个识别到的时间字段和上一个是相连的 @author kexm
			rPointer -= 1
			temp[rPointer] = temp[rPointer] + m.String() // 则把下一个识别到的时间字段加到上一个时间字段去
		} else {
			temp = append(temp, m.String())
		}
		pos = append(pos, m.Index)
		length = append(length, m.Length)
		endLine = m.Index + m.Length
		rPointer += 1
		m, _ = n.pattern.FindNextMatch(m)
	}
	var res []TimeUnit
	// 时间上下文： 前一个识别出来的时间会是下一个时间的上下文，用于处理：周六3点到5点这样的多个时间的识别，第二个5点应识别到是周六的。
	tpCtx := DefaultTimePoint
	idx := 0
	for idx < rPointer {
		unit := NewTimeUnit(temp[idx], pos[idx], length[idx], n, tpCtx)
		idx += 1
		if unit.ts.IsZero() {
			continue
		}
		res = append(res, *unit)
		tpCtx = unit.tp
	}
	res = n.filterTimeUnits(res)
	return res
}

// filterTimeUnits 过滤空时间点
func (n *TimeNormalizer) filterTimeUnits(units []TimeUnit) []TimeUnit {
	var res []TimeUnit
	for _, tu := range units {
		if tu.ts.IsZero() {
			continue
		}
		res = append(res, tu)
	}
	return res
}
