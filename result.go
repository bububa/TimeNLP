package timenlp

import "time"

// ResultType 返回值类型
type ResultType string

const (
	DELTA     ResultType = "delta"     // 相对时间
	SPAN      ResultType = "span"      // 时间段
	TIMESTAMP ResultType = "timestamp" // 时间点
)

// ResultPoint 返回值包含时间点
type ResultPoint struct {
	Time   time.Time // 时间
	Pos    int       `json:"pos,omitempty"`    // 文字位置
	Length int       `json:"length,omitempty"` // 文字长度
}

// Result 返回值
type Result struct {
	NormalizedString string        `json:"normalized_string,omitempty"` // 标准化后字符串
	Type             ResultType    `json:"type,omitempty"`              // 返回类型
	Points           []ResultPoint `json:"points,omitempty"`            // 时间点
}
