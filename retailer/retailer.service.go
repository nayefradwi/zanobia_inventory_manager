package retailer

import "github.com/nayefradwi/zanobia_inventory_manager/common"

type IRetailerService interface{}

type RetailerService struct {
	repo         IRetailerRepository
	batchService IRetailerBatchService
}

func NewRetailerService(repo IRetailerRepository, batchService IRetailerBatchService) *RetailerService {
	return &RetailerService{
		repo,
		batchService,
	}
}

func (s *RetailerService) CreateRetailer(retailer Retailer) error {
	return nil
}

func (s *RetailerService) AddRetailerContacts(retailerId int, contacts []RetailerContact) error {
	return nil
}

func (s *RetailerService) AddRetailerContactInfo(retailerId int, contact RetailerContact) error {
	return nil
}

func (s *RetailerService) GetRetailers(params common.PaginationParams) ([]Retailer, error) {
	return nil, nil
}

func (s *RetailerService) GetRetailer(id int) (Retailer, error) {
	return Retailer{}, nil
}

func (s *RetailerService) RemoveRetailerContactInfo(id int) error {
	return nil
}

func (s *RetailerService) RemoveRetailer(id int) error {
	return nil
}

func (s *RetailerService) UpdateRetailer(retailer Retailer) error {
	return nil
}
