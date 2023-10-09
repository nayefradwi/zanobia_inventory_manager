package product

import "github.com/nayefradwi/zanobia_inventory_manager/common"

type ProductInput struct {
	ProductBase
	ExpiresInDays  int       `json:"expiresInDays"`
	StandardUnitId *int      `json:"standardUnitId,omitempty"`
	Price          float64   `json:"price"`
	Variants       []Variant `json:"variants,omitempty"`
}

type ProductBase struct {
	Id          *int    `json:"id"`
	Name        *string `json:"name"`
	Description string  `json:"description"`
	Image       string  `json:"image,omitempty"`
	IsArchived  bool    `json:"isArchived"`
	CategoryId  *int    `json:"categoryId,omitempty"`
}

type Product struct {
	ProductBase
	Category        *Category        `json:"category,omitempty"`
	Options         []Variant        `json:"options,omitempty"`
	ProductVariants []ProductVariant `json:"productVariants,omitempty"`
}

type ProductVariantBase struct {
	Id             *int     `json:"id"`
	ProductId      *int     `json:"productId"`
	Name           string   `json:"name"`
	Sku            string   `json:"sku"`
	Image          string   `json:"image,omitempty"`
	Price          float64  `json:"price"`
	WidthInCm      *float64 `json:"widthInCm,omitempty"`
	HeightInCm     *float64 `json:"heightInCm,omitempty"`
	DepthInCm      *float64 `json:"depthInCm,omitempty"`
	WeightInG      *float64 `json:"weightInG,omitempty"`
	StandardUnitId *int     `json:"standardUnitId,omitempty"`
	IsArchived     bool     `json:"isArchived"`
	IsDefault      bool     `json:"isDefault"`
	ExpiresInDays  int      `json:"expiresInDays"`
}

type ProductVariant struct {
	ProductVariantBase
	StandardUnit *Unit `json:"standardUnit,omitempty"`
}

func (p ProductInput) GenerateProductVariantsFromOptions() []ProductVariant {
	productVariants := []ProductVariant{}
	if len(p.Variants) == 0 {
		productVariants = append(productVariants, p.GenerateProductVariantForNoOptionsProduct())
		return productVariants
	}
	for index, option := range p.Variants {
		productVariant := p.GenerateProductVariantFromOption(option, index == 0)
		productVariants = append(productVariants, productVariant)
	}
	return productVariants
}

func (p ProductInput) GenerateProductVariantForNoOptionsProduct() ProductVariant {
	uuid, _ := common.GenerateUuid()
	return ProductVariant{
		ProductVariantBase: ProductVariantBase{
			Price:          p.Price,
			IsArchived:     p.IsArchived,
			IsDefault:      true,
			Image:          p.Image,
			StandardUnitId: p.StandardUnitId,
			ExpiresInDays:  p.ExpiresInDays,
			Name:           "normal",
			Sku:            uuid,
		},
	}
}

func (p ProductInput) GenerateProductVariantFromOption(option Variant, isDefault bool) ProductVariant {
	uuid, _ := common.GenerateUuid()
	name := option.GenerateNameFromValues()
	return ProductVariant{
		ProductVariantBase: ProductVariantBase{
			Price:          p.Price,
			IsArchived:     p.IsArchived,
			IsDefault:      true,
			Image:          p.Image,
			StandardUnitId: p.StandardUnitId,
			ExpiresInDays:  p.ExpiresInDays,
			Name:           name,
			Sku:            uuid,
		},
	}
}
