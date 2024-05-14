package retailer

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"go.uber.org/zap"
)

type IRetailerRepository interface {
	CreateRetailer(ctx context.Context, retailer Retailer) error
	AddRetailerContacts(ctx context.Context, retailerId int, contacts []RetailerContact) error
	AddRetailerContactInfo(ctx context.Context, retailerId int, contact RetailerContact) error
	GetRetailers(ctx context.Context, params common.PaginationParams) ([]Retailer, error)
	GetRetailer(ctx context.Context, retailerId int) (Retailer, error)
	RemoveRetailerContactInfo(ctx context.Context, contactId int) error
	RemoveRetailer(ctx context.Context, retailerId int) error
	RemoveRetailerTranslations(ctx context.Context, retailerId int) error
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
	sql := `INSERT INTO retailers (lat, lng) VALUES ($1, $2) RETURNING id`
	op := common.GetOperator(ctx, r.Pool)
	row := op.QueryRow(ctx, sql, retailer.Lat, retailer.Lng)
	var id int
	err := row.Scan(&id)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to create retailer", zap.Error(err))
		return 0, common.NewBadRequestFromMessage("Failed to create retailer")
	}
	return id, nil
}

func (r *RetailerRepo) translateRetialer(ctx context.Context, id int, retailer Retailer, languageCode string) error {
	sql := `INSERT INTO retailer_translations (retailer_id, language_code, name) VALUES ($1, $2, $3)`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, id, languageCode, retailer.Name)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to create retailer translation", zap.Error(err))
		return common.NewBadRequestFromMessage("Failed to create retailer translation")
	}
	return nil
}

func (r *RetailerRepo) AddRetailerContacts(ctx context.Context, retailerId int, contacts []RetailerContact) error {
	err := common.RunWithTransaction(ctx, r.Pool, func(ctx context.Context, tx pgx.Tx) error {
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
	sql := `INSERT INTO retailer_contact_info (retailer_id, email, phone, website) VALUES ($1, $2, $3, $4) RETURNING id`
	op := common.GetOperator(ctx, r.Pool)
	row := op.QueryRow(ctx, sql, retailerId, contact.Email, contact.Phone, contact.Website)
	var id int
	err := row.Scan(&id)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to create retailer contact info", zap.Error(err))
		return 0, common.NewBadRequestFromMessage("Failed to create retailer contact info")
	}
	return id, nil
}

func (r *RetailerRepo) translateRetailerContactInfo(ctx context.Context, id int, contact RetailerContact, languageCode string) error {
	sql := `INSERT INTO retailer_contact_info_translations (retailer_contact_info_id, language_code, name, position) VALUES ($1, $2, $3, $4)`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, id, languageCode, contact.Name, contact.Position)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to create retailer contact info translation", zap.Error(err))
		return common.NewBadRequestFromMessage("Failed to create retailer contact info translation")
	}
	return nil
}

func (r *RetailerRepo) GetRetailers(ctx context.Context, params common.PaginationParams) ([]Retailer, error) {
	op := common.GetOperator(ctx, r.Pool)
	langCode := common.GetLanguageParam(ctx)
	rows, err := common.NewPaginationQueryBuilder(
		`select r.id, r.lat, r.lng, rtx.name from retailers r
		 join retailer_translations rtx on rtx.retailer_id = r.id
		`,
		[]string{"r.id desc"},
	).
		WithOperator(op).
		WithConditions([]string{"rtx.language_code = $1"}).
		WithParams(params).
		WithCursorKeys([]string{"r.id"}).
		WithCompareSymbols("<", "<=", ">").
		Build().
		Query(ctx, langCode)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to get retailers", zap.Error(err))
		return []Retailer{}, common.NewBadRequestFromMessage("Failed to get retailers")
	}
	defer rows.Close()
	var retailers []Retailer
	for rows.Next() {
		var retailer Retailer
		err := rows.Scan(&retailer.Id, &retailer.Lat, &retailer.Lng, &retailer.Name)
		if err != nil {
			common.LoggerFromCtx(ctx).Error("Failed to get retailers", zap.Error(err))
			return []Retailer{}, common.NewBadRequestFromMessage("Failed to get retailers")
		}
		retailers = append(retailers, retailer)
	}
	return retailers, nil
}

func (r *RetailerRepo) GetRetailer(ctx context.Context, retailerId int) (Retailer, error) {
	sql := `
	SELECT r.id, r.lat, r.lng, rtx.name, rcti.email, rcti.phone, rcti.website,
	rctitx.name, rctitx.position, rcti.id
	FROM retailers r
	JOIN retailer_translations rtx ON rtx.retailer_id = r.id
	RIGHT JOIN retailer_contact_info rcti ON rcti.retailer_id = r.id
	JOIN retailer_contact_info_translations rctitx ON rctitx.retailer_contact_info_id = rcti.id
	WHERE r.id = $1 AND rtx.language_code = $2 AND rctitx.language_code = $2
	`
	op := common.GetOperator(ctx, r.Pool)
	langCode := common.GetLanguageParam(ctx)
	rows, err := op.Query(ctx, sql, retailerId, langCode)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to get retailer", zap.Error(err))
		return Retailer{}, common.NewBadRequestFromMessage("Failed to get retailer")
	}
	defer rows.Close()
	var retailer Retailer
	for rows.Next() {
		var contact RetailerContact
		var contactEmail, contactWebsite *string
		err := rows.Scan(
			&retailer.Id, &retailer.Lat, &retailer.Lng,
			&retailer.Name, &contactEmail,
			&contact.Phone, &contactWebsite,
			&contact.Name,
			&contact.Position,
			&contact.Id,
		)
		if err != nil {
			common.LoggerFromCtx(ctx).Error("Failed to get retailer", zap.Error(err))
			return Retailer{}, common.NewBadRequestFromMessage("Failed to get retailer")
		}
		if contactEmail != nil {
			contact.Email = *contactEmail
		}
		if contactWebsite != nil {
			contact.Website = *contactWebsite
		}
		retailer.Contacts = append(retailer.Contacts, contact)
	}
	return retailer, nil
}

func (r *RetailerRepo) RemoveRetailerContactInfo(ctx context.Context, contactInfoId int) error {
	err := common.RunWithTransaction(ctx, r.Pool, func(ctx context.Context, tx pgx.Tx) error {
		if err := r.removeRetailerContactInfoTranslation(ctx, contactInfoId); err != nil {
			return err
		}
		if err := r.removeRetailerContactInfo(ctx, contactInfoId); err != nil {
			return err
		}
		return nil
	})
	return err
}

func (r *RetailerRepo) removeRetailerContactInfo(ctx context.Context, contactInfoId int) error {
	sql := `DELETE FROM retailer_contact_info WHERE id = $1`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, contactInfoId)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to remove retailer contact info", zap.Error(err))
		return common.NewBadRequestFromMessage("Failed to remove retailer contact info")
	}
	return nil
}

func (r *RetailerRepo) removeRetailerContactInfoTranslation(ctx context.Context, contactInfoId int) error {
	sql := `DELETE FROM retailer_contact_info_translations WHERE retailer_contact_info_id = $1`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, contactInfoId)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to remove retailer contact info translation", zap.Error(err))
		return common.NewBadRequestFromMessage("Failed to remove retailer contact info translation")
	}
	return nil
}

func (r *RetailerRepo) RemoveRetailer(ctx context.Context, retailerId int) error {
	sql := `DELETE FROM retailers WHERE id = $1`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, retailerId)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to remove retailer", zap.Error(err))
		return common.NewBadRequestFromMessage("Failed to remove retailer")
	}
	return nil
}

func (r *RetailerRepo) RemoveRetailerTranslations(ctx context.Context, retailerId int) error {
	sql := `DELETE FROM retailer_translations WHERE retailer_id = $1`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, retailerId)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to remove retailer translations", zap.Error(err))
		return common.NewBadRequestFromMessage("Failed to remove retailer translations")
	}
	return nil
}

func (r *RetailerRepo) UpdateRetailer(ctx context.Context, retailer Retailer) error {
	return common.RunWithTransaction(ctx, r.Pool, func(ctx context.Context, tx pgx.Tx) error {
		if err := r.updateRetailerLatLng(ctx, *retailer.Id, retailer.Lat, retailer.Lng); err != nil {
			return err
		}
		if err := r.updateRetailerName(ctx, *retailer.Id, retailer.Name); err != nil {
			return err
		}
		return nil
	})
}

func (r *RetailerRepo) updateRetailerLatLng(ctx context.Context, retailerId int, lat float64, lng float64) error {
	sql := `UPDATE retailers SET lat = $1, lng = $2 WHERE id = $3`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, lat, lng, retailerId)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to update retailer", zap.Error(err))
		return common.NewBadRequestFromMessage("Failed to update retailer")
	}
	return nil
}

func (r *RetailerRepo) updateRetailerName(ctx context.Context, retailerId int, name string) error {
	sql := `UPDATE retailer_translations SET name = $1 WHERE retailer_id = $2`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, name, retailerId)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to update retailer", zap.Error(err))
		return common.NewBadRequestFromMessage("Failed to update retailer")
	}
	return nil
}

func (r *RetailerRepo) RemoveAllContactsOfRetailer(ctx context.Context, retailerId int) error {
	if err := r.removeAllContactInfoTranslationsOfRetailer(ctx, retailerId); err != nil {
		return err
	}
	if err := r.removeAllContactsOfRetailer(ctx, retailerId); err != nil {
		return err
	}
	return nil
}

func (r RetailerRepo) removeAllContactsOfRetailer(ctx context.Context, retailerId int) error {
	sql := `DELETE FROM retailer_contact_info WHERE retailer_id = $1`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, retailerId)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to remove retailer contact info", zap.Error(err))
		return common.NewBadRequestFromMessage("Failed to remove retailer contact info")
	}
	return nil
}

func (r RetailerRepo) removeAllContactInfoTranslationsOfRetailer(ctx context.Context, retailerId int) error {
	sql := `
	DELETE FROM retailer_contact_info_translations USING retailer_contact_info
	WHERE retailer_contact_info_translations.retailer_contact_info_id = retailer_contact_info.id
	AND 
	retailer_contact_info.retailer_id = $1`
	op := common.GetOperator(ctx, r.Pool)
	_, err := op.Exec(ctx, sql, retailerId)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("Failed to remove retailer contact info translation", zap.Error(err))
		return common.NewBadRequestFromMessage("Failed to remove retailer contact info translation")
	}
	return nil
}
