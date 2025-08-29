package shared

type PendingCollection struct {
	Amount         int32 `json:"amount" validate:"required"`
	CollectionDate Date  `json:"collection_date" validate:"required"`
}
