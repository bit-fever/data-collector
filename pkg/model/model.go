package model

type SymbolData struct {
	Code     string  `json:"code,omitempty"`
	Name     string  `json:"name"`
	MakerFee float64 `json:"maker_fee"`
	TakerFee float64 `json:"taker_fee"`
	ApiKey   string  `json:"api_key"`
	Secret   string  `json:"secret"`
	Test     bool    `json:"test"`
}
