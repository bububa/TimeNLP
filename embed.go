package timenlp

import (
	// embed data
	_ "embed"
)

//go:embed resource/regex.txt
var embedPattern string

//go:embed resource/holi_solar.json
var embedHoliSolar []byte

//go:embed resource/holi_lunar.json
var embedHoliLunar []byte
