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
	Category        *Category        `json:"category,omitempty"`
	Options         []Variant        `json:"options,omitempty"`
	ProductVariants []ProductVariant `json:"productVariants,omitempty"`
}

type ProductInput struct {
	ProductBase
	ExpiresInDays                int       `json:"expiresInDays"`
	StandardUnitId               *int      `json:"standardUnitId,omitempty"`
	Price                        float64   `json:"price"`
	Variants                     []Variant `json:"variants,omitempty"`
	ProductVariantsLookupByValue map[string][]ProductVariant
	ProductVariants              []ProductVariant
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

func generateCrossProductOfValueNames(variants []Variant) []VariantValue {
	crossProduct := make([]VariantValue, 0)
	if len(variants) == 2 {
		return getCrossProduct(variants[0].Values, variants[1].Values)
	}
	i := 0
	for i < len(variants) {
		variant1 := variants[i]
		variants2 := variants[i+1]
		initialProduct := getCrossProduct(variant1.Values, variants2.Values)
		variants[i+1] = Variant{Values: initialProduct}
		variants = variants[i+1:]
		i++
	}
	return crossProduct
}

func getCrossProduct(values1 []VariantValue, values2 []VariantValue) []VariantValue {
	crossProduct := make([]VariantValue, 0)
	for _, value1 := range values1 {
		for _, value2 := range values2 {
			crossProduct = append(crossProduct, VariantValue{
				Value: value1.Value + "_" + value2.Value,
			})
		}
	}
	return crossProduct
}

func generateProductVariantsLookupByValue(productVariants []ProductVariant) map[string][]ProductVariant {
	// TODO fill
	return map[string][]ProductVariant{}
}

func (p ProductInput) GenerateProductDetails() ProductInput {
	productVariants := p.generateProductVariants()
	p.ProductVariants = productVariants
	p.ProductVariantsLookupByValue = generateProductVariantsLookupByValue(productVariants)
	return p
}

// assume you have packaging, weight, and flavor
// packaging: 1, 2, 3
// weight: a, b
// flavor: @, #
// then you will have 12 variants
// [1_a_@, 1_a_#, 1_b_@, 1_b_#, 2_a_@, 2_a_#, 2_b_@, 2_b_#, 3_a_@, 3_a_#, 3_b_@, 3_b_#]
// the first variant is the default one
func (p ProductInput) generateProductVariants() []ProductVariant {
	if len(p.Variants) == 0 {
		return []ProductVariant{p.createProductVariant(VariantValue{Value: "normal"}, true)}
	}
	if len(p.Variants) == 1 {
		return p.createProductVariantsFromOneOption()
	}
	crossProductOfNames := generateCrossProductOfValueNames(p.Variants)
	productVariants := make([]ProductVariant, 0)
	for index, value := range crossProductOfNames {
		productVariant := p.createProductVariant(value, index == 0)
		productVariants = append(productVariants, productVariant)
	}
	return productVariants
}

func (p ProductInput) createProductVariantsFromOneOption() []ProductVariant {
	productVariants := make([]ProductVariant, 0)
	for index, value := range p.Variants[0].Values {
		productVariant := p.createProductVariant(value, index == 0)
		productVariants = append(productVariants, productVariant)
	}
	return productVariants
}

func (p ProductInput) createProductVariant(value VariantValue, isDefault bool) ProductVariant {
	uuid, _ := common.GenerateUuid()
	return ProductVariant{
		ProductVariantBase: ProductVariantBase{
			Price:          p.Price,
			IsArchived:     p.IsArchived,
			IsDefault:      true,
			Image:          p.Image,
			StandardUnitId: p.StandardUnitId,
			ExpiresInDays:  p.ExpiresInDays,
			Name:           value.Value,
			Sku:            uuid,
		},
	}
}
