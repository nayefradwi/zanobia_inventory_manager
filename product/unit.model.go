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
	ConversionFactor float64 `json:"conversionFactor"`
}

type UnitConversionInput struct {
	UnitName           string  `json:"unitName"`
	ConversionUnitName string  `json:"conversionUnitName"`
	ConversionFactor   float64 `json:"conversionFactor"`
}

type ConvertUnitInput struct {
	UnitId           *int    `json:"unitId"`
	ConversionUnitId *int    `json:"conversionUnitId"`
	Quantity         float64 `json:"quantity"`
}

type ConvertUnitOutput struct {
	Unit     Unit    `json:"unit"`
	Quantity float64 `json:"quantity"`
}
