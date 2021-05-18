package timenlp

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/dlclark/regexp2"
)

type StringPreHandler struct{}

// 该方法删除一字符串中所有匹配某一规则字串
// 可用于清理一个字符串中的空白符和语气助词
// :param target: 待处理字符串
// :param rules: 删除规则
// :return: 清理工作完成后的字符串
func (s StringPreHandler) DelKeyword(target string, rules string) string {
	pattern := regexp.MustCompile(rules)
	return pattern.ReplaceAllString(target, "")
}

// 该方法可以将字符串中所有的用汉字表示的数字转化为用阿拉伯数字表示的数字
// 如"这里有一千两百个人，六百零五个来自中国"可以转化为
// "这里有1200个人，605个来自中国"
// 此外添加支持了部分不规则表达方法
// 如两万零六百五可转化为20650
// 两百一十四和两百十四都可以转化为214
// 一六零加一五八可以转化为160+158
// 该方法目前支持的正确转化范围是0-99999999
// 该功能模块具有良好的复用性
// :param target: 待转化的字符串
// :return: 转化完毕后的字符串
func (s StringPreHandler) NumberTranslator(target string) string {
	{
		pattern := regexp2.MustCompile("[一二两三四五六七八九123456789]万[一二两三四五六七八九123456789](?!(千|百|十))", 0)
		var match *regexp2.Match
		for {
			if match == nil {
				match, _ = pattern.FindStringMatch(target)
			} else {
				match, _ = pattern.FindNextMatch(match)
			}
			if match == nil {
				break
			}
			matchedString := match.String()
			parts := s.filterStringSlice(strings.Split(matchedString, "万"), "")
			var num int64
			if len(parts) == 2 {
				num += s.WordToNum(parts[0])*10000 + s.WordToNum(parts[1])*1000
			}
			target = strings.Replace(target, matchedString, strconv.FormatInt(num, 10), 1)
		}
	}
	{
		pattern := regexp2.MustCompile("[一二两三四五六七八九123456789]千[一二两三四五六七八九123456789](?!(百|十))", 0)
		var match *regexp2.Match
		for {
			if match == nil {
				match, _ = pattern.FindStringMatch(target)
			} else {
				match, _ = pattern.FindNextMatch(match)
			}
			if match == nil {
				break
			}
			matchedString := match.String()
			parts := s.filterStringSlice(strings.Split(matchedString, "千"), "")
			var num int64
			if len(parts) == 2 {
				num += s.WordToNum(parts[0])*1000 + s.WordToNum(parts[1])*100
			}
			target = strings.Replace(target, matchedString, strconv.FormatInt(num, 10), 1)
		}
	}
	{
		pattern := regexp2.MustCompile("[一二两三四五六七八九123456789]百[一二两三四五六七八九123456789](?!十)", 0)
		var match *regexp2.Match
		for {
			if match == nil {
				match, _ = pattern.FindStringMatch(target)
			} else {
				match, _ = pattern.FindNextMatch(match)
			}
			if match == nil {
				break
			}
			matchedString := match.String()
			parts := s.filterStringSlice(strings.Split(matchedString, "百"), "")
			var num int64
			if len(parts) == 2 {
				num += s.WordToNum(parts[0])*100 + s.WordToNum(parts[1])*10
			}
			target = strings.Replace(target, matchedString, strconv.FormatInt(num, 10), 1)
		}
	}
	{
		pattern := regexp.MustCompile("[零一二两三四五六七八九]")
		if match := pattern.FindAllString(target, -1); match != nil {
			for _, m := range match {
				num := s.WordToNum(m)
				target = strings.Replace(target, m, strconv.FormatInt(num, 10), 1)
			}
		}
	}
	{
		pattern := regexp2.MustCompile("(?<=(周|星期))[末天日]", 0)
		var match *regexp2.Match
		for {
			if match == nil {
				match, _ = pattern.FindStringMatch(target)
			} else {
				match, _ = pattern.FindNextMatch(match)
			}
			if match == nil {
				break
			}
			matchedString := match.String()
			num := s.WordToNum(matchedString)
			target = strings.Replace(target, matchedString, strconv.FormatInt(num, 10), 1)
		}
	}
	{
		pattern := regexp2.MustCompile("(?<!(周|星期))0?[0-9]?十[0-9]?", 0)
		var match *regexp2.Match
		for {
			if match == nil {
				match, _ = pattern.FindStringMatch(target)
			} else {
				match, _ = pattern.FindNextMatch(match)
			}
			if match == nil {
				break
			}
			matchedString := match.String()
			parts := strings.Split(matchedString, "十")
			ten, _ := strconv.ParseInt(parts[0], 10, 64)
			if ten == 0 {
				ten = 1
			}
			unit, _ := strconv.ParseInt(parts[1], 10, 64)
			num := ten*10 + unit
			target = strings.Replace(target, matchedString, strconv.FormatInt(num, 10), 1)
		}
	}
	{
		pattern := regexp.MustCompile("0?[1-9]百[0-9]?[0-9]?")
		if match := pattern.FindAllString(target, -1); match != nil {
			for _, m := range match {
				parts := s.filterStringSlice(strings.Split(m, "百"), "")
				var num int64
				if len(parts) == 1 {
					hundred, _ := strconv.ParseInt(parts[0], 10, 64)
					num += hundred * 100
				} else if len(parts) == 2 {
					hundred, _ := strconv.ParseInt(parts[0], 10, 64)
					num += hundred * 100
					unit, _ := strconv.ParseInt(parts[1], 10, 64)
					num += unit
				}
				target = strings.Replace(target, m, strconv.FormatInt(num, 10), 1)
			}
		}
	}
	{
		pattern := regexp.MustCompile("0?[1-9]千[0-9]?[0-9]?[0-9]?")
		if match := pattern.FindAllString(target, -1); match != nil {
			for _, m := range match {
				parts := s.filterStringSlice(strings.Split(m, "千"), "")
				var num int64
				if len(parts) == 1 {
					thousand, _ := strconv.ParseInt(parts[0], 10, 64)
					num += thousand * 1000
				} else if len(parts) == 2 {
					thousand, _ := strconv.ParseInt(parts[0], 10, 64)
					num += thousand * 1000
					unit, _ := strconv.ParseInt(parts[1], 10, 64)
					num += unit
				}
				target = strings.Replace(target, m, strconv.FormatInt(num, 10), 1)
			}
		}
	}
	{
		pattern := regexp.MustCompile("[0-9]+万[0-9]?[0-9]?[0-9]?[0-9]?")
		if match := pattern.FindAllString(target, -1); match != nil {
			for _, m := range match {
				parts := s.filterStringSlice(strings.Split(m, "万"), "")
				var num int64
				if len(parts) == 1 {
					tenThousand, _ := strconv.ParseInt(parts[0], 10, 64)
					num += tenThousand * 10000
				} else if len(parts) == 2 {
					tenThousand, _ := strconv.ParseInt(parts[0], 10, 64)
					num += tenThousand * 10000
					unit, _ := strconv.ParseInt(parts[1], 10, 64)
					num += unit
				}
				target = strings.Replace(target, m, strconv.FormatInt(num, 10), 1)
			}
		}
	}
	return target
}

func (s StringPreHandler) filterStringSlice(arr []string, f string) []string {
	var ret []string
	for _, a := range arr {
		if a == f {
			continue
		}
		ret = append(ret, a)
	}
	return ret
}

// 方法numberTranslator的辅助方法，可将[零-九]正确翻译为[0-9]
// :param s: 大写数字
// :return: 对应的整形数，如果不是数字返回-1
func (s StringPreHandler) WordToNum(str string) int64 {
	switch str {
	case "零", "0":
		return 0
	case "一", "1":
		return 1
	case "二", "两", "2":
		return 2
	case "三", "3":
		return 3
	case "四", "4":
		return 4
	case "五", "5":
		return 5
	case "六", "6":
		return 6
	case "七", "7", "天", "日", "末":
		return 7
	case "八", "8":
		return 8
	case "九", "9":
		return 9
	}
	return -1
}
