package shared

type PendingCollection struct {
	Amount         int  `json:"amount" validate:"required"`
	CollectionDate Date `json:"collection_date" validate:"required"`
}
