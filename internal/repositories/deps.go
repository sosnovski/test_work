package repositories

import (
	"context"

	"github.com/jackc/pgtype/pgxtype"
)

type Querier interface {
	Conn(ctx context.Context) pgxtype.Querier
}
