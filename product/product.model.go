package product

import (
	"strconv"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type ProductOption struct {
	Id     *int                 `json:"id,omitempty"`
	Name   string               `json:"name"`
	Values []ProductOptionValue `json:"values"`
}

type ProductOptionValue struct {
	Id    *int   `json:"id,omitempty"`
	Value string `json:"value"`
}

type AddVariantValueInput struct {
	ProductOptionId int    `json:"productOptionId"`
	Value           string `json:"value"`
}

func (a AddVariantValueInput) ToProductOptionValue() ProductOptionValue {
	return ProductOptionValue{
		Value: a.Value,
	}
}

type ProductBase struct {
	Id           *int    `json:"id"`
	Name         *string `json:"name"`
	Description  string  `json:"description"`
	Image        string  `json:"image,omitempty"`
	IsArchived   bool    `json:"isArchived"`
	CategoryId   *int    `json:"categoryId,omitempty"`
	IsIngredient bool    `json:"isIngredient"`
}

func (p ProductBase) GetCursorValue() []string {
	return []string{strconv.Itoa(*p.Id)}
}

type Product struct {
	ProductBase
	Category        *Category                `json:"category,omitempty"`
	Options         map[string]ProductOption `json:"options,omitempty"`
	ProductVariants []ProductVariant         `json:"productVariants,omitempty"`
}

type ProductInput struct {
	ProductBase
	ExpiresInDays         int             `json:"expiresInDays"`
	StandardUnitId        *int            `json:"standardUnitId,omitempty"`
	Price                 float64         `json:"price"`
	Options               []ProductOption `json:"options,omitempty"`
	ProductVariants       []ProductVariant
	DefaultProductVariant ProductVariant
	DefaultOptionValues   []ProductOptionValue
}

type ProductVariantBase struct {
	Id             *int     `json:"id,omitempty"`
	ProductId      *int     `json:"productId,omitempty"`
	Name           string   `json:"name"`
	Sku            string   `json:"sku,omitempty"`
	Image          string   `json:"image,omitempty"`
	Price          float64  `json:"price,omitempty"`
	WidthInCm      *float64 `json:"widthInCm,omitempty"`
	HeightInCm     *float64 `json:"heightInCm,omitempty"`
	DepthInCm      *float64 `json:"depthInCm,omitempty"`
	WeightInG      *float64 `json:"weightInG,omitempty"`
	StandardUnitId *int     `json:"standardUnitId,omitempty"`
	IsArchived     bool     `json:"isArchived"`
	IsDefault      bool     `json:"isDefault"`
	ExpiresInDays  int      `json:"expiresInDays,omitempty"`
}

type ProductVariant struct {
	ProductVariantBase
	ProductName  string   `json:"productName"`
	TotalCost    float64  `json:"totalCost,omitempty"`
	Recipes      []Recipe `json:"recipes,omitempty"`
	StandardUnit *Unit    `json:"standardUnit,omitempty"`
}

type ProductVariantInput struct {
	ProductVariant ProductVariant `json:"productVariant"`
	OptionValueIds []int          `json:"optionsValueIds,omitempty"`
	OptionValues   []ProductOptionValue
}

func (p ProductInput) GenerateProductDetails() ProductInput {
	productVariant, defaultValues := p.generateProductVariant()
	p.ProductVariants = []ProductVariant{productVariant}
	p.DefaultProductVariant = productVariant
	p.DefaultOptionValues = defaultValues
	return p
}

// assume you have packaging, weight, and flavor
// the first variant will have packaging[0]_weight[0]_flavor[0]
// the reset can be added in any order by the user
// the selected variant values will be mapped to the product to limit the options
// the first variant is the default one
func (p ProductInput) generateProductVariant() (ProductVariant, []ProductOptionValue) {
	if len(p.Options) == 0 {
		return p.createProductVariant("normal", true), []ProductOptionValue{}
	}
	optionValues := make([]ProductOptionValue, 0)
	name := ""
	for _, variant := range p.Options {
		optionValues = append(optionValues, variant.Values[0])
		name += variant.Values[0].Value + "_"
	}
	name = name[:len(name)-1]
	productVariant := p.createProductVariant(name, true)
	return productVariant, optionValues
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

func GenerateName(values []ProductOptionValue) string {
	name := ""
	for _, value := range values {
		name += value.Value + "_"
	}
	return name[:len(name)-1]
}
