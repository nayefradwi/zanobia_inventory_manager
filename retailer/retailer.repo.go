package retailer

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type IRetailerRepository interface {
	CreateRetailer(ctx context.Context, retailer Retailer) error
	AddRetailerContacts(ctx context.Context, retailerId int, contacts []RetailerContact) error
	AddRetailerContactInfo(ctx context.Context, retailerId int, contact RetailerContact) error
	GetRetailers(ctx context.Context, params common.PaginationParams) ([]Retailer, error)
	GetRetailer(ctx context.Context, retailerId int) (Retailer, error)
	RemoveRetailerContactInfo(ctx context.Context, contactId int) error
	RemoveRetailer(ctx context.Context, retailerId int) error
	UpdateRetailer(ctx context.Context, retailer Retailer) error
	RemoveAllContactsOfRetailer(ctx context.Context, retailerId int) error
}

type RetailerRepo struct {
	*pgxpool.Pool
}

func NewRetailerRepository(db *pgxpool.Pool) *RetailerRepo {
	return &RetailerRepo{
		db,
	}
}

func (r *RetailerRepo) CreateRetailer(ctx context.Context, retailer Retailer) error {
	err := common.RunWithTransaction(ctx, r.Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		id, err := r.insertRetailer(ctx, retailer)
		if err != nil {
			return err
		}
		if err := r.translateRetialer(ctx, id, retailer, common.DefaultLang); err != nil {
			return err
		}
		if err := r.addRetailerContacts(ctx, id, retailer.Contacts); err != nil {
			return err
		}
		return nil
	})
	return err
}

func (r *RetailerRepo) insertRetailer(ctx context.Context, retailer Retailer) (int, error) {
	// TODO fill
	return 0, nil
}

func (r *RetailerRepo) translateRetialer(ctx context.Context, id int, retailer Retailer, languageCode string) error {
	// TODO fill
	return nil
}

func (r *RetailerRepo) AddRetailerContacts(ctx context.Context, retailerId int, contacts []RetailerContact) error {
	err := common.RunWithTransaction(ctx, r.Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		return r.addRetailerContacts(ctx, retailerId, contacts)
	})
	return err
}

func (r *RetailerRepo) addRetailerContacts(ctx context.Context, retailerId int, contacts []RetailerContact) error {
	for _, contact := range contacts {
		err := r.addRetailerContactInfo(ctx, retailerId, contact)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *RetailerRepo) AddRetailerContactInfo(ctx context.Context, retailerId int, contact RetailerContact) error {
	err := common.RunWithTransaction(ctx, r.Pool, func(ctx context.Context, tx pgx.Tx) error {
		ctx = common.SetOperator(ctx, tx)
		return r.addRetailerContactInfo(ctx, retailerId, contact)
	})
	return err
}

func (r *RetailerRepo) addRetailerContactInfo(ctx context.Context, retailerId int, contact RetailerContact) error {
	id, err := r.insertRetailerContactInfo(ctx, retailerId, contact)
	if err != nil {
		return err
	}
	if err := r.translateRetailerContactInfo(ctx, id, contact, common.DefaultLang); err != nil {
		return err
	}
	return nil
}

func (r *RetailerRepo) insertRetailerContactInfo(ctx context.Context, retailerId int, contact RetailerContact) (int, error) {
	// TODO fill
	return 0, nil
}

func (r *RetailerRepo) translateRetailerContactInfo(ctx context.Context, id int, contact RetailerContact, languageCode string) error {
	// TODO fill
	return nil
}

func (r *RetailerRepo) GetRetailers(ctx context.Context, params common.PaginationParams) ([]Retailer, error) {
	// TODO fill
	return []Retailer{}, nil
}

func (r *RetailerRepo) GetRetailer(ctx context.Context, retailerId int) (Retailer, error) {
	// TODO fill
	return Retailer{}, nil
}

func (r *RetailerRepo) RemoveRetailerContactInfo(ctx context.Context, contactInfoId int) error {
	// TODO fill
	return nil
}

func (r *RetailerRepo) RemoveRetailer(ctx context.Context, retailerId int) error {
	// TODO fill
	return nil
}

func (r *RetailerRepo) UpdateRetailer(ctx context.Context, retailer Retailer) error {
	// TODO fill
	return nil
}

func (r *RetailerRepo) RemoveAllContactsOfRetailer(ctx context.Context, retailerId int) error {
	// TODO fill
	return nil
}
