package timenlp

import "time"

// ResultType 返回值类型
type ResultType string

const (
	// DELTA 相对时间
	DELTA ResultType = "delta"
	// SPAN 时间段
	SPAN ResultType = "span"
	// TIMESTAMP 时间点
	TIMESTAMP ResultType = "timestamp"
)

// ResultPoint 返回值包含时间点
type ResultPoint struct {
	// Time 时间
	Time time.Time
	// Pos 文字位置
	Pos int `json:"pos,omitempty"`
	// Length 文字长度
	Length int `json:"length,omitempty"`
}

// Result 返回值
type Result struct {
	// NormalizedString 标准化后字符串
	NormalizedString string `json:"normalized_string,omitempty"`
	// Type 返回类型
	Type ResultType `json:"type,omitempty"`
	// Points 时间点
	Points []ResultPoint `json:"points,omitempty"`
}
