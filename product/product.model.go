package product

type ProductBase struct {
	Id             *int     `json:"id"`
	Name           *string  `json:"name"`
	Description    string   `json:"description"`
	Image          *string  `json:"image,omitempty"`
	Price          float64  `json:"price"`
	WidthInCm      *float64 `json:"widthInCm,omitempty"`
	HeightInCm     *float64 `json:"heightInCm,omitempty"`
	DepthInCm      *float64 `json:"depthInCm,omitempty"`
	WeightInG      *float64 `json:"weightInG,omitempty"`
	StandardUnitId *int     `json:"standardUnitId,omitempty"`
	CategoryId     *int     `json:"categoryId,omitempty"`
	IsArchived     bool     `json:"isArchived"`
	ExpiresInDays  int      `json:"expiresInDays"`
}

type Product struct {
	ProductBase
	StandardUnit *Unit     `json:"standardUnit,omitempty"`
	Category     *Category `json:"category,omitempty"`
	// TODO missing recipe
}
