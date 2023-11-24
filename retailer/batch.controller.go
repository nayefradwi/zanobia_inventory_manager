package retailer

type RetailerBatchController struct {
	service IRetailerBatchService
}

func NewRetailerBatchController(service IRetailerBatchService) *RetailerBatchController {
	return &RetailerBatchController{
		service,
	}
}
