package product

import (
	"sort"
	"strconv"
	"strings"

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

type ProductOptionInput struct {
	ProductId int                       `json:"productId"`
	Name      string                    `json:"optionName"`
	Values    []ProductOptionValueInput `json:"values"`
}

type ProductOptionValueInput struct {
	Value     string `json:"value"`
	IsDefault bool   `json:"isDefault"`
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
	productVariants := make([]ProductVariant, 1)
	skuLookup := make(map[string]OptionValueSet)
	allValues := make([]OptionValueSet, 1)
	for _, option := range p.Options {
		firstValue := option.Values[0]
		allValues[0] = append(allValues[0], firstValue)
	}
	name := GenerateName(allValues[0])
	productVariants[0] = p.createProductVariant(name, true)
	return productVariants, skuLookup
}

func (p ProductInput) createProductVariant(value string, isDefault bool) ProductVariant {
	uuid, _ := common.GenerateUuid()
	return ProductVariant{
		ProductVariantBase: ProductVariantBase{
			Price:          p.Price,
			IsArchived:     p.IsArchived,
			IsDefault:      isDefault,
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

func (pvar ProductVariant) AddValueToName(value string) ProductVariant {
	variantNameSplit := strings.Split(pvar.Name, "_")
	variantNameSplit = append(variantNameSplit, value)
	values := make([]ProductOptionValue, len(variantNameSplit))
	for i, value := range variantNameSplit {
		values[i] = ProductOptionValue{
			Value: value,
		}
	}
	newName := GenerateName(values)
	pvar.Name = newName
	return pvar
}
