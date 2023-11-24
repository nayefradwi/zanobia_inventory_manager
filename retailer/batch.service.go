package retailer

type IRetailerBatchService interface{}

type RetailerBatchService struct {
	repo IRetailerBatchRepository
}

func NewRetailerBatchService(repo IRetailerBatchRepository) *RetailerBatchService {
	return &RetailerBatchService{
		repo,
	}
}
