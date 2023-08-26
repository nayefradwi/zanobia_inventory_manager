package product

type Unit struct {
	Id     *int   `json:"id,omitempty"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

type UnitConversion struct {
	Id               *int    `json:"id,omitempty"`
	UnitId           *int    `json:"unitId"`
	ConversionUnitId *int    `json:"conversionUnitId"`
	ConversionFactor float32 `json:"conversionFactor"`
}

type UnitConversionInput struct {
	UnitName           string  `json:"unitName"`
	ConversionUnitName string  `json:"conversionUnitName"`
	ConversionFactor   float32 `json:"conversionFactor"`
}
