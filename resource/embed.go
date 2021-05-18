package resource

import (
	_ "embed"
)

//go:embed regex.txt
var Pattern string

//go:embed holi_solar.json
var HoliSolar []byte

//go:embed holi_lunar.json
var HoliLunar []byte
