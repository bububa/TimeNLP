package timenlp

import "time"

type ResultType string

const (
	DELTA     ResultType = "delta"
	SPAN      ResultType = "span"
	TIMESTAMP ResultType = "timestamp"
)

type ResultPoint struct {
	Time   time.Time
	Pos    int `json:"pos,omitempty"`
	Length int `json:"length,omitempty"`
}

type Result struct {
	NormalizedString string        `json:"normalized_string,omitempty"`
	Type             ResultType    `json:"type,omitempty"`
	Points           []ResultPoint `json:"points,omitempty"`
}
