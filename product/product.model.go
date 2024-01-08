package product

import (
	"sort"
	"strconv"

	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/unit"
)

type UpdateSkuInput struct {
	OldSku string `json:"oldSku"`
	NewSku string `json:"newSku"`
}

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
	skuOptionValuesLookup map[string]OptionValueSet
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
	ProductName  string     `json:"productName"`
	IsIngredient bool       `json:"isIngredient"`
	TotalCost    float64    `json:"totalCost,omitempty"`
	Recipes      []Recipe   `json:"recipes,omitempty"`
	StandardUnit *unit.Unit `json:"standardUnit,omitempty"`
}

type ProductVariantUpdate struct {
	Id         int      `json:"id,omitempty"`
	Price      float64  `json:"price,omitempty"`
	WidthInCm  *float64 `json:"widthInCm,omitempty"`
	HeightInCm *float64 `json:"heightInCm,omitempty"`
	DepthInCm  *float64 `json:"depthInCm,omitempty"`
	WeightInG  *float64 `json:"weightInG,omitempty"`
	IsArchived bool     `json:"isArchived"`
}

type ProductVariantInput struct {
	ProductVariant ProductVariant `json:"productVariant"`
	OptionValueIds []int          `json:"optionsValueIds,omitempty"`
	OptionValues   []ProductOptionValue
}

func (p ProductInput) GenerateProductDetails() ProductInput {
	productVariants, valuesLookup := p.generateProductVariant()
	p.ProductVariants = productVariants
	p.skuOptionValuesLookup = valuesLookup
	return p
}

type OptionValueSet []ProductOptionValue

func (p ProductInput) createNormalProductVariant() ([]ProductVariant, map[string]OptionValueSet) {
	return []ProductVariant{p.createProductVariant("normal", true)}, map[string]OptionValueSet{}
}

// assume you have packaging, weight, and flavor
// the first variant will have packaging[0]_weight[0]_flavor[0]
// the reset can be added in any order by the user
// the selected variant values will be mapped to the product to limit the options
// the first variant is the default one
func (p ProductInput) generateProductVariant() ([]ProductVariant, map[string]OptionValueSet) {
	if len(p.Options) == 0 {
		return p.createNormalProductVariant()
	}
	return p.createCrossProductOfVariants()
}

func (p ProductInput) createCrossProductOfVariants() ([]ProductVariant, map[string]OptionValueSet) {
	cartesianLength := 1
	for _, options := range p.Options {
		cartesianLength *= len(options.Values)
	}
	productVariants := make([]ProductVariant, cartesianLength)
	allValues := make([]OptionValueSet, cartesianLength)
	skuLookup := make(map[string]OptionValueSet)
	for i := 0; i < cartesianLength; i++ {
		allValues[i] = make([]ProductOptionValue, len(p.Options))
		for optionIndex, option := range p.Options {
			allValues[i][optionIndex] = option.Values[i%len(option.Values)]
		}
		name := GenerateName(allValues[i])
		productVariants[i] = p.createProductVariant(name, i == 0)
		skuLookup[productVariants[i].Sku] = allValues[i]
	}
	return productVariants, skuLookup
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
	SortOptionValues(values)
	for _, value := range values {
		name += value.Value + "_"
	}
	return name[:len(name)-1]
}

func SortOptionValues(values []ProductOptionValue) {
	sort.Slice(values, func(i, j int) bool {
		valueI, valueJ := values[i].Value, values[j].Value
		if z, err := strconv.Atoi(valueI); err == nil {
			if y, err := strconv.Atoi(valueJ); err == nil {
				return y < z
			}
			return true
		}
		return valueJ > valueI
	})
}

func (pvar ProductVariant) GetCursorValue() []string {
	return []string{strconv.Itoa(*pvar.Id)}
}
