package product

type IIngredientService interface{}

type IngredientService struct {
	repo IIngredientRepository
}

func NewIngredientService(repo IIngredientRepository) IIngredientService {
	return &IngredientService{
		repo,
	}
}
