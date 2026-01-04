package patterns

// 边界字符定义
const (
	SplitPattern  = `★★／/_&（）\s\-\.\[\]\(\)`
	BoundaryStart = `[` + SplitPattern + `]`
	BoundaryEnd   = `(?=[` + SplitPattern + `])`
)

// ChineseNumberMap 中文数字到阿拉伯数字的映射
var ChineseNumberMap = map[string]int{
	"一": 1, "二": 2, "三": 3, "四": 4, "五": 5,
	"六": 6, "七": 7, "八": 8, "九": 9, "十": 10,
}

// ChineseNumberUpperMap 大写中文数字到阿拉伯数字的映射
var ChineseNumberUpperMap = map[string]int{
	"零": 0, "壹": 1, "贰": 2, "叁": 3, "肆": 4,
	"伍": 5, "陆": 6, "柒": 7, "捌": 8, "玖": 9,
}

// RomanNumbers 罗马数字到阿拉伯数字的映射
var RomanNumbers = map[string]int{
	"I": 1, "II": 2, "III": 3, "IV": 4, "V": 5,
}
