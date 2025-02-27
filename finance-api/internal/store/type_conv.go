package store

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"math"
)

func ToInt4(dest *pgtype.Int4, i any) error {
	switch i.(type) {
	case int:
		v := i.(int)
		if v < math.MinInt32 || v > math.MaxInt32 {
			return fmt.Errorf("cannot scan %T", i)
		}
		dest.Int32 = int32(v)
		dest.Valid = true
	case int32:
		v := i.(int32)
		if v < 0 {
			return fmt.Errorf("cannot scan %T", i)
		}
		dest.Int32 = v
		dest.Valid = true
	default:
		return dest.Scan(i)
	}
	return nil
}

func toInt32(i int) (int32, error) {
	if i < math.MinInt32 || i > math.MaxInt32 {
		return 0, fmt.Errorf("cannot scan %T", i)
	}
	return int32(i), nil
}
