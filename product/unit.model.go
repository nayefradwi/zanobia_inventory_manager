package product

type Unit struct {
	Id     *int   `json:"id,omitempty"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

type UnitConversion struct {
	Id               *int    `json:"id,omitempty"`
	ToUnitId         *int    `json:"toUnitId"`
	FromUnitId       *int    `json:"fromUnitId"`
	ConversionFactor float64 `json:"conversionFactor"`
}

type UnitConversionInput struct {
	ToUnitName       string  `json:"toUnitName"`
	FromUnitName     string  `json:"fromUnitName"`
	ConversionFactor float64 `json:"conversionFactor"`
}

type ConvertUnitInput struct {
	ToUnitId   *int    `json:"toUnitId"`
	FromUnitId *int    `json:"fromUnitId"`
	Quantity   float64 `json:"quantity"`
}

type ConvertUnitOutput struct {
	Unit     Unit    `json:"unit"`
	Quantity float64 `json:"quantity"`
}
