package retailer

type RetailerController struct {
	service IRetailerService
}

func NewRetailController(service IRetailerService) *RetailerController {
	return &RetailerController{
		service,
	}
}
