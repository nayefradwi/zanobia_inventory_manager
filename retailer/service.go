package retailer

type IRetailerService interface{}

type RetailerService struct {
	repo IRetailerRepository
}

func NewRetailerService(repo IRetailerRepository) *RetailerService {
	return &RetailerService{
		repo,
	}
}
