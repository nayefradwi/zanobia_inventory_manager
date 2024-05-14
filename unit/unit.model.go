package unit

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

const (
	Grams       = "grams"
	Kilograms   = "kilograms"
	Milliliters = "milliliters"
	Liters      = "liters"
	Piece       = "piece"
	Tablespoon  = "tablespoon"
	Teaspoon    = "teaspoon"
	Jar         = "jar"
	Carton      = "carton"
	Bottle      = "bottle"
	Box         = "box"
)

var initialUnits = []Unit{
	{
		Name:   Grams,
		Symbol: "g",
	},
	{
		Name:   Kilograms,
		Symbol: "kg",
	},
	{
		Name:   Milliliters,
		Symbol: "ml",
	},
	{
		Name:   Liters,
		Symbol: "L",
	},
	{
		Name:   Piece,
		Symbol: "pc",
	},
	{
		Name:   Tablespoon,
		Symbol: "tbsp",
	},
	{
		Name:   Teaspoon,
		Symbol: "tsp",
	},

	{
		Name:   Jar,
		Symbol: "jar",
	},

	{
		Name:   Carton,
		Symbol: "carton",
	},

	{
		Name:   Bottle,
		Symbol: "bottle",
	},

	{
		Name:   Box,
		Symbol: "box",
	},
}

var initialConversions = []UnitConversionInput{
	{
		ToUnitName:       Kilograms,
		FromUnitName:     Grams,
		ConversionFactor: 0.001,
	},
	{
		ToUnitName:       Grams,
		FromUnitName:     Kilograms,
		ConversionFactor: 1000,
	},
	{
		ToUnitName:       Liters,
		FromUnitName:     Milliliters,
		ConversionFactor: 0.001,
	},
	{
		ToUnitName:       Milliliters,
		FromUnitName:     Liters,
		ConversionFactor: 1000,
	},

	{
		ToUnitName:       Grams,
		FromUnitName:     Tablespoon,
		ConversionFactor: 15,
	},
	{
		ToUnitName:       Grams,
		FromUnitName:     Teaspoon,
		ConversionFactor: 5,
	},
	{
		ToUnitName:       Milliliters,
		FromUnitName:     Tablespoon,
		ConversionFactor: 15,
	},
	{
		ToUnitName:       Milliliters,
		FromUnitName:     Teaspoon,
		ConversionFactor: 5,
	},
	{
		ToUnitName:       Liters,
		FromUnitName:     Tablespoon,
		ConversionFactor: 0.015,
	},
	{
		ToUnitName:       Liters,
		FromUnitName:     Teaspoon,
		ConversionFactor: 0.005,
	},
	{
		ToUnitName:       Kilograms,
		FromUnitName:     Tablespoon,
		ConversionFactor: 0.015,
	},
	{
		ToUnitName:       Kilograms,
		FromUnitName:     Teaspoon,
		ConversionFactor: 0.005,
	},
	{
		ToUnitName:       Tablespoon,
		FromUnitName:     Grams,
		ConversionFactor: 0.67,
	},
	{
		ToUnitName:       Teaspoon,
		FromUnitName:     Grams,
		ConversionFactor: 0.2,
	},
	{
		ToUnitName:       Tablespoon,
		FromUnitName:     Milliliters,
		ConversionFactor: 0.67,
	},
	{
		ToUnitName:       Teaspoon,
		FromUnitName:     Milliliters,
		ConversionFactor: 0.2,
	},
	{
		ToUnitName:       Tablespoon,
		FromUnitName:     Liters,
		ConversionFactor: 670,
	},
	{
		ToUnitName:       Teaspoon,
		FromUnitName:     Liters,
		ConversionFactor: 200,
	},
	{
		ToUnitName:       Tablespoon,
		FromUnitName:     Kilograms,
		ConversionFactor: 670,
	},
	{
		ToUnitName:       Teaspoon,
		FromUnitName:     Kilograms,
		ConversionFactor: 200,
	},
	{
		ToUnitName:       Liters,
		FromUnitName:     Kilograms,
		ConversionFactor: 1,
	},
	{
		ToUnitName:       Kilograms,
		FromUnitName:     Liters,
		ConversionFactor: 1,
	},
	{
		ToUnitName:       Milliliters,
		FromUnitName:     Kilograms,
		ConversionFactor: 1000,
	},
	{
		ToUnitName:       Kilograms,
		FromUnitName:     Milliliters,
		ConversionFactor: 0.001,
	},
	{
		ToUnitName:       Milliliters,
		FromUnitName:     Grams,
		ConversionFactor: 1,
	},
	{
		ToUnitName:       Grams,
		FromUnitName:     Milliliters,
		ConversionFactor: 1,
	},
	{
		ToUnitName:       Liters,
		FromUnitName:     Grams,
		ConversionFactor: 0.001,
	},
	{
		ToUnitName:       Grams,
		FromUnitName:     Liters,
		ConversionFactor: 1000,
	},
}
