package store

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"math"
)

func ToInt4(dest *pgtype.Int4, i any) error {
	switch i := i.(type) {
	case int:
		if i < math.MinInt32 || i > math.MaxInt32 {
			return fmt.Errorf("cannot scan %T", i)
		}
		dest.Int32 = int32(i)
		dest.Valid = i != 0
	case int32:
		if i < 0 {
			return fmt.Errorf("cannot scan %T", i)
		}
		dest.Int32 = i
		dest.Valid = i != 0
	default:
		return dest.Scan(i)
	}
	return nil
}
