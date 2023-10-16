package product

import (
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

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
	Category        *Category          `json:"category,omitempty"`
	Options         map[string]Variant `json:"options,omitempty"`
	ProductVariants []ProductVariant   `json:"productVariants,omitempty"`
}

type ProductInput struct {
	ProductBase
	ExpiresInDays         int       `json:"expiresInDays"`
	StandardUnitId        *int      `json:"standardUnitId,omitempty"`
	Price                 float64   `json:"price"`
	Variants              []Variant `json:"variants,omitempty"`
	ProductVariants       []ProductVariant
	DefaultProductVariant ProductVariant
	DefaultValues         []VariantValue
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
	ProductName  string   `json:"productName"`
	Recipes      []Recipe `json:"recipes,omitempty"`
	StandardUnit *Unit    `json:"standardUnit,omitempty"`
}

type ProductVariantInput struct {
	ProductVariant ProductVariant `json:"productVariant"`
	VariantValues  []VariantValue `json:"variantValues,omitempty"`
}

func (p ProductInput) GenerateProductDetails() ProductInput {
	productVariant, defaultValues := p.generateProductVariant()
	p.ProductVariants = []ProductVariant{productVariant}
	p.DefaultProductVariant = productVariant
	p.DefaultValues = defaultValues
	return p
}

// assume you have packaging, weight, and flavor
// the first variant will have packaging[0]_weight[0]_flavor[0]
// the reset can be added in any order by the user
// the selected variant values will be mapped to the product to limit the options
// the first variant is the default one
func (p ProductInput) generateProductVariant() (ProductVariant, []VariantValue) {
	if len(p.Variants) == 0 {
		return p.createProductVariant("normal", true), []VariantValue{}
	}
	variantValues := make([]VariantValue, 0)
	name := ""
	for _, variant := range p.Variants {
		variantValues = append(variantValues, variant.Values[0])
		name += variant.Values[0].Value + "_"
	}
	name = name[:len(name)-1]
	productVariant := p.createProductVariant(name, true)
	return productVariant, variantValues
}

func (p ProductInput) createProductVariant(value string, isDefault bool) ProductVariant {
	uuid, _ := common.GenerateUuid()
	return ProductVariant{
		ProductVariantBase: ProductVariantBase{
			Price:          p.Price,
			IsArchived:     p.IsArchived,
			IsDefault:      true,
			Image:          p.Image,
			StandardUnitId: p.StandardUnitId,
			ExpiresInDays:  p.ExpiresInDays,
			Name:           value,
			Sku:            uuid,
		},
	}
}
