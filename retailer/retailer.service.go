package retailer

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IRetailerService interface {
	CreateRetailer(ctx context.Context, retailer Retailer) error
	AddRetailerContacts(ctx context.Context, retailerId int, contacts []RetailerContact) error
	AddRetailerContactInfo(ctx context.Context, retailerId int, contact RetailerContact) error
	GetRetailers(ctx context.Context) (common.PaginatedResponse[Retailer], error)
	GetRetailer(ctx context.Context, id int) (Retailer, error)
	RemoveRetailerContactInfo(ctx context.Context, id int) error
	RemoveRetailer(ctx context.Context, id int) error
	UpdateRetailer(ctx context.Context, retailer Retailer) error
}

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

func (s *RetailerService) CreateRetailer(ctx context.Context, retailer Retailer) error {
	if err := ValidateRetailer(retailer); err != nil {
		return err
	}
	return s.repo.CreateRetailer(ctx, retailer)
}

func (s *RetailerService) AddRetailerContacts(ctx context.Context, retailerId int, contacts []RetailerContact) error {
	if err := ValidateRetailerContacts(contacts); err != nil {
		return err
	}
	return s.repo.AddRetailerContacts(ctx, retailerId, contacts)
}

func (s *RetailerService) AddRetailerContactInfo(ctx context.Context, retailerId int, contact RetailerContact) error {
	if err := ValidateRetailerContact(contact); err != nil {
		return err
	}
	return s.repo.AddRetailerContactInfo(ctx, retailerId, contact)
}

func (s *RetailerService) GetRetailers(ctx context.Context) (common.PaginatedResponse[Retailer], error) {
	params := common.GetPaginationParams(ctx)
	retailers, err := s.repo.GetRetailers(ctx, params)
	if err != nil {
		return common.CreateEmptyPaginatedResponse[Retailer](params.PageSize), err
	}
	if len(retailers) == 0 {
		return common.CreateEmptyPaginatedResponse[Retailer](params.PageSize), nil
	}
	first, last := retailers[0], retailers[len(retailers)-1]
	paginatedResponse := common.CreatePaginatedResponse[Retailer](
		params.PageSize,
		last,
		first,
		retailers,
	)
	return paginatedResponse, nil
}

func (s *RetailerService) GetRetailer(ctx context.Context, id int) (Retailer, error) {
	return s.repo.GetRetailer(ctx, id)
}

func (s *RetailerService) RemoveRetailerContactInfo(ctx context.Context, id int) error {
	return s.repo.RemoveRetailerContactInfo(ctx, id)
}

func (s *RetailerService) RemoveRetailer(ctx context.Context, id int) error {
	return common.RunWithTransaction(ctx, s.repo.(*RetailerRepo).Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		if err := s.batchService.DeleteBatchesOfRetailer(ctx, id); err != nil {
			return err
		}
		if err := s.repo.RemoveAllContactsOfRetailer(ctx, id); err != nil {
			return err
		}
		if err := s.repo.RemoveRetailer(ctx, id); err != nil {
			return err
		}
		return nil
	})
}

func (s *RetailerService) UpdateRetailer(ctx context.Context, retailer Retailer) error {
	if err := ValidateRetailer(retailer); err != nil {
		return err
	}
	return s.repo.UpdateRetailer(ctx, retailer)
}
