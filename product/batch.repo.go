package product

import "github.com/jackc/pgx/v4/pgxpool"

type IBatchRepository interface{}

type BatchRepository struct {
	*pgxpool.Pool
}

func NewBatchRepository(pool *pgxpool.Pool) *BatchRepository {
	return &BatchRepository{pool}
}

// Create batch
// Edit batch quantity
// Get batch base by product id and expiration date
// Get batches paginated sorted by expiration date
